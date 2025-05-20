#!/usr/bin/env bash

# No true constants so using read only just to prevent bugs
current_date=$(date +"%Y-%m-%d")
readonly edm_abbrev="EDM"
readonly scoreboard_endpt="https://api-web.nhle.com/v1/scoreboard/${edm_abbrev}/now"
readonly schedule_now_enpt="https://api-web.nhle.com/v1/schedule/${current_date}"

# Format to match the date format in the response
#echo "The current date is ${current_date}"

get_games_today(){
    schedule_now_data=$(curl -sX GET $schedule_now_enpt)

    game_week=$(echo $schedule_now_data | jq -c '.gameWeek[]')
    # Not using quotes with the echo flattens the json causing issues...
    echo "$game_week" | while read -r day; do
        date=$(echo "$day" | jq -r '.date')
        if [[ "$date" == "$current_date" ]]; then
            games=$(echo "$day" | jq -c '.games[]')
            echo "$games"
        fi
    done
}



games_today=$(get_games_today)


#json_data=$(curl -X GET $scoreboard_endpt)
# Piping to jq in this case just pretty prints the data
#echo $json_data | jq .
