#!/usr/bin/env bash

# Survival RPG: A terminal-based survival/RPG game written in pure Bash.
# To play, make this file executable and run: ./ascii_mud.sh

ascii_mud() {
    # Shell compatibility (Zsh emulate ksh)
    if [ -n "$ZSH_VERSION" ]; then
        emulate -L ksh
    fi

    # --- Configuration ---
    local WIDTH=40
    local HEIGHT=20
    local PLAYER_CHAR="@"
    local MONSTER_CHAR="M"
    local TREE_CHAR="T"
    local WATER_CHAR="~"
    local EMPTY_CHAR="."
    local GAME_LOOP_SLEEP=0.1

    # Player state
    local player_x=$((WIDTH/2))
    local player_y=$((HEIGHT/2))
    local health=100
    local hunger=100
    local exp=0
    local level=1

    # World arrays
    local map=()
    local monsters_x=() monsters_y=() monsters_hp=()

    # Trap and cleanup
    cleanup() {
        stty icanon echo
        tput cnorm
        tput sgr0
        echo
    }
    trap cleanup EXIT

    # Initialize terminal
    stty -icanon -echo
    tput civis
    clear

    # Generate map
    for ((y=0; y<HEIGHT; y++)); do
        for ((x=0; x<WIDTH; x++)); do
            if (( RANDOM % 100 < 10 )); then
                map[y*WIDTH+x]="${TREE_CHAR}"
            elif (( RANDOM % 100 < 5 )); then
                map[y*WIDTH+x]="${WATER_CHAR}"
            else
                map[y*WIDTH+x]="${EMPTY_CHAR}"
            fi
        done
    done

    # Spawn initial monsters
    for ((i=0; i<5; i++)); do
        monsters_x+=( $((RANDOM % WIDTH)) )
        monsters_y+=( $((RANDOM % HEIGHT)) )
        monsters_hp+=( $((RANDOM % 20 + 10)) )
    done

    draw() {
        tput cup 0 0
        # Status
        printf "HP:%3d  Hunger:%3d  EXP:%3d  LEVEL:%d\n" "$health" "$hunger" "$exp" "$level"
        # Map
        for ((y=0; y<HEIGHT; y++)); do
            for ((x=0; x<WIDTH; x++)); do
                if (( x == player_x && y == player_y )); then
                    printf "%s" "$PLAYER_CHAR"
                else
                    local drawn=0
                    for idx in "${!monsters_x[@]}"; do
                        if (( monsters_x[idx] == x && monsters_y[idx] == y )); then
                            printf "%s" "$MONSTER_CHAR"
                            drawn=1; break
                        fi
                    done
                    if (( drawn == 0 )); then
                        printf "%s" "${map[y*WIDTH+x]}"
                    fi
                fi
            done
            printf "\n"
        done
        printf "Controls: WASD move, f=attack, q=quit\n"
    }

    # Game loop
    local action
    while (( health > 0 && hunger > 0 )); do
        draw
        # Input
        IFS= read -rsn1 action
        case "$action" in
            w) (( player_y > 0 )) && player_y=$((player_y-1));;
            s) (( player_y < HEIGHT-1 )) && player_y=$((player_y+1));;
            a) (( player_x > 0 )) && player_x=$((player_x-1));;
            d) (( player_x < WIDTH-1 )) && player_x=$((player_x+1));;
            f)
                # Attack adjacent monster
                for idx in "${!monsters_x[@]}"; do
                    dx=$(( monsters_x[idx] - player_x ))
                    dy=$(( monsters_y[idx] - player_y ))
                    if (( (dx*dx + dy*dy) == 1 )); then
                        monsters_hp[idx]=$(( monsters_hp[idx] - (RANDOM % 20 + 5) ))
                        if (( monsters_hp[idx] <= 0 )); then
                            exp=$(( exp + 10 ))
                            # remove monster
                            unset monsters_x[idx] monsters_y[idx] monsters_hp[idx]
                        fi
                        break
                    fi
                done
                ;;
            q) break;;
        esac

        # Hunger and health decay
        hunger=$(( hunger - 1 ))
        if (( hunger % 20 == 0 )); then
            health=$(( health - 5 ))
        fi

        # Monster movement (random)
        for idx in "${!monsters_x[@]}"; do
            if (( RANDOM % 2 == 0 )); then
                dir=$((RANDOM % 4))
                case $dir in
                    0) dx=1; dy=0;; 1) dx=-1; dy=0;;
                    2) dx=0; dy=1;; 3) dx=0; dy=-1;;
                esac
                newx=$(( monsters_x[idx] + dx ))
                newy=$(( monsters_y[idx] + dy ))
                if (( newx >=0 && newx < WIDTH && newy >=0 && newy < HEIGHT )); then
                    monsters_x[idx]=$newx; monsters_y[idx]=$newy
                fi
            fi
            # Monster attack if on player
            if (( monsters_x[idx] == player_x && monsters_y[idx] == player_y )); then
                health=$(( health - (RANDOM % 10 + 5) ))
            fi
        done

        # Level up
        if (( exp >= level*50 )); then
            exp=$(( exp - level*50 ))
            level=$(( level + 1 ))
            health=$(( health + 20 ))
            hunger=$(( hunger + 20 ))
        fi

        sleep "$GAME_LOOP_SLEEP"
    done

    # Game over
    clear
    echo "========================="
    echo "       GAME OVER         "
    echo " You reached level ${level}   "
    echo " Remaining HP: ${health}   "
    echo " Hunger: ${hunger}        "
    echo "========================="
    cleanup
}

