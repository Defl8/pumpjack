# pumpjack
Script for fetching information about live and upcoming Oilers hockey games. Originally intended to be a waybar custom module.

# Road Map (Ordered by priority)
- [x] ~~D.R.Y.ify the codebase.~~ Did the best I could with what I know. Probably better ways to do this.
- [x] ~~Add better exit codes and handling for when *curl* command fails.~~
- [ ] Return the data as json from the script? (Still not sure about this.) 
    - This would be so that it could be formatted by waybar or other things.
    - Turns out this is probably really simple with *printf*
- [x] ~~Functionality for all teams in the NHL.~~ Might make an internal map so the full team name can be passed and not just the abbreviation.
- [ ] Add shootout support.
    - Meh
