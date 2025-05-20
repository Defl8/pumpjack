#!/usr/bin/env bash

# No true constants so using read only just to prevent bugs
current_date=$(date +"%Y-%m-%d")
readonly edm_abbrev="EDM"
readonly scoreboard_endpt="https://api-web.nhle.com/v1/scoreboard/${edm_abbrev}/now"
readonly schedule_now_endpt="https://api-web.nhle.com/v1/schedule/${current_date}"

# Format to match the date format in the response
#echo "The current date is ${current_date}"

get_games_today(){
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
    readonly boxscore_endpt="https://api-web.nhle.com/v1/gamecenter/${game_id}/boxscore"
    local game_info=$(curl -sX GET $boxscore_endpt)
    local game_state=$(echo "$game_info" | jq -r ".gameState")
    if [[ "$game_state" == "LIVE" ]]; then
        return 0
    fi
    return 1
}

games_today=$(get_games_today)

game_id=0 # used in check_if_playing_today 
if check_if_playing_today "$games_today"; then
    # TODO: Remove later
    if check_if_game_live "$game_id"; then
    #if check_if_game_live 2024030311; then
        echo "The game is live"
    else
        echo "Game is not live"
    fi
else
    # TODO: Remove later, whole if block
    if check_if_game_live 2024030311; then
        echo "The game is live"
    else
        echo "Game is not live"
    fi
fi
