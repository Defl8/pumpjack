#!/usr/bin/env bash

# No true constants so using read only just to prevent bugs
current_date=$(date +"%Y-%m-%d")
readonly edm_abbrev="EDM"

# Format to match the date format in the response
#echo "The current date is ${current_date}"

get_games_today(){
    readonly schedule_now_endpt="https://api-web.nhle.com/v1/schedule/${current_date}"
    local schedule_now_data=$(curl -sX GET $schedule_now_endpt)

    local game_week=$(echo $schedule_now_data | jq -c '.gameWeek[]')
    # Not using quotes with the echo flattens the json causing issues...
    echo "$game_week" | while read -r day; do
        local date=$(echo "$day" | jq -r '.date')
        if [[ "$date" == "$current_date" ]]; then
            local games=$(echo "$day" | jq -c '.games[]')
            echo "$games"
        fi
    done
}

check_if_playing_today(){
    local games_today="$1"
    while read -r game; do
        local game_state=$(echo "$game" | jq -c '.gameState')

        local away_team=$(echo "$game" | jq -c '.awayTeam.abbrev')
        local home_team=$(echo "$game" | jq -c '.homeTeam.abbrev')
        if [[ "$home_team" == "$edm_abbrev" || "$away_team" == "$edm_abbrev" ]]; then
            game_id=$(echo "$game" | jq -c '.id')
            return 0 # Game is being played today
        fi
    return 1 # Game is not being played today

# while read runs in subshell, this triple arrow is a here string that keeps
# the loop in the same shell so that I can use the return value
    done <<< "$games_today" 
}

check_if_game_live(){
    local game_id="$1"
    local boxscore_endpt="https://api-web.nhle.com/v1/gamecenter/${game_id}/boxscore"
    local game_info=$(curl -sX GET $boxscore_endpt)
    local game_state=$(echo "$game_info" | jq -r ".gameState")
    # "FUT" means future
    if [[ "$game_state" != "OFF" || "$game_state" != "FUT" ]]; then
        return 0
    fi
    return 1
}

format_live(){
    local game_id="$1"
    local boxscore_endpt="https://api-web.nhle.com/v1/gamecenter/${game_id}/boxscore"
    local game_info=$(curl -sX GET $boxscore_endpt)
    local game_state=$(echo "$game_info" | jq -r ".gameState")

    local away_team=$(echo "$game_info" | jq -r '.awayTeam.abbrev')
    local home_team=$(echo "$game_info" | jq -r '.homeTeam.abbrev')
    local away_score=$(echo "$game_info" | jq -r '.awayTeam.score')
    local home_score=$(echo "$game_info" | jq -r '.homeTeam.score')

    local time_remaining=$(echo "$game_info" | jq -r '.clock.timeRemaining')
    local period_num=$(echo "$game_info" | jq -r '.periodDescriptor.number')
    local num_denom="nd"

    local period="${period}${period_num}"
    if (( period_num == 3)); then
        period="${period}rd"
    elif (( period_num == 4)); then
        period="OT"
    fi

    if [[ "$game_state" != "FINAL" ]]; then
        echo "$away_team $away_score - $period $time_remaining - $home_score $home_team"
    else
        if [[ "$period" != "OT" ]]; then
            echo "$away_team $away_score - FINAL - $home_score $home_team"
        else
            echo "$away_team $away_score - FINAL/OT - $home_score $home_team"
        fi
    fi
}

format_later_today(){
    local game_id="$1"
    local boxscore_endpt="https://api-web.nhle.com/v1/gamecenter/${game_id}/boxscore"
    local game_info=$(curl -sX GET $boxscore_endpt)

    local away_team=$(echo "$game_info" | jq -r '.awayTeam.abbrev')
    local home_team=$(echo "$game_info" | jq -r '.homeTeam.abbrev')
    local game_time_utc=$(echo "$game_info" | jq -r '.startTimeUTC')
    local game_time_mst=$(date -d "$game_time_utc" +"Today @ %H:%M")
    echo "$game_time_mst | $away_team @ $home_team"
}

get_next_game(){
    local schedule_week_endpt="https://api-web.nhle.com/v1/club-schedule/${edm_abbrev}/week/${current_date}" 
    local week_info=$(curl -sX GET "$schedule_week_endpt")

    # The first element in the list of games for the week is always the one
    # closest to the current date
    local next_game=$(echo "$week_info" | jq -c '.games[0]')
    echo "$next_game"
}

format_next_game(){
    local next_game="$1"
    local away_team=$(echo "$next_game" | jq -r '.awayTeam.abbrev')
    local home_team=$(echo "$next_game" | jq -r '.homeTeam.abbrev')

    local game_time_utc=$(echo "$next_game" | jq -r '.startTimeUTC')
    local game_time_mst=$(date -d "$game_time_utc" +"%a @ %H:%M")
    if [[ -z "$next_game" ]]; then
        echo "No games scheduled"
    else
        echo "$game_time_mst | $away_team @ $home_team"
    fi
}

games_today=$(get_games_today)
game_id=0 # used in check_if_playing_today 

if check_if_playing_today "$games_today"; then
    if check_if_game_live "$game_id"; then
        live_output=$(format_live "$game_id")
        echo "$live_output"
    else
        later_today_output=$(format_later_today "$game_id")
        echo "$later_today_output"
    fi
else
    next_game=$(get_next_game)
    next_game_output=$(format_next_game "$next_game")
    echo "$next_game_output"
fi
