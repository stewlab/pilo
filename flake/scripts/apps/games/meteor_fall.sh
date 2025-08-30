#!/usr/bin/env bash

# Meteor Fall: A simple terminal-based action game written in pure Bash.
# This script defines the `meteor_fall` function. To play, source this file
# and then run the command: meteor_fall

meteor_fall() {
    # --- Shell Compatibility ---
    # If in Zsh, emulate ksh for the duration of the function. This provides
    # maximum compatibility with Bash-like array handling and built-in commands.
    if [ -n "$ZSH_VERSION" ]; then
        emulate -L ksh
    fi

    # --- Game Configuration ---
    local WIDTH=40
    local HEIGHT=20
    local PLAYER_CHAR="▲"
    local METEOR_CHAR="✭"
    local PICKUP_CHAR="♦"
    local BULLET_CHAR="|"
    local EMPTY_CHAR=" "
    local GAME_LOOP_SLEEP=0.02

    # --- Game State Variables ---
    local player_pos=$((WIDTH / 2))
    local score=0
    local level=1
    local score_to_next_level=100
    local game_over=0
    local tick=0
    # Rate at which objects update or spawn. Lower is faster/more frequent.
    local meteor_update_rate=4
    local pickup_update_rate=8 # Pickups now fall a bit slower
    local bullet_update_rate=1 # Bullets are fast
    local meteor_spawn_rate=12
    local pickup_spawn_rate=60 # Pickups are a bit rarer

    # Arrays to store object properties
    local meteor_x=() meteor_y=() meteor_dx=()
    local pickup_x=() pickup_y=()
    local bullet_x=() bullet_y=()

    # --- Setup and Cleanup ---
    cleanup() {
        stty icanon echo
        tput cnorm
        tput sgr0
        echo
    }
    # Set the trap only for the duration of this function
    trap cleanup EXIT

    # --- Drawing and Spawning Functions ---
    draw_screen() {
        local screen_buffer
        screen_buffer="Meteor Fall | Score: $score | Level: $level\n"
        for ((i=0; i<WIDTH+2; i++)); do screen_buffer+="─"; done; screen_buffer+="\n"
        for ((y=0; y<HEIGHT; y++)); do
            screen_buffer+="│"
            for ((x=0; x<WIDTH; x++)); do
                screen_buffer+="${screen[y*WIDTH+x]}"
            done
            screen_buffer+="│\n"
        done
        for ((i=0; i<WIDTH+2; i++)); do screen_buffer+="─"; done; screen_buffer+="\n"
        screen_buffer+="Controls: [a] Left | [d] Right | [space] Shoot | [q] Quit\n"
        tput cup 0 0
        printf "%b" "$screen_buffer"
    }

    update_meteors() {
        local new_x=() new_y=() new_dx=()
        for i in "${!meteor_x[@]}"; do
            local next_y=$((meteor_y[i] + 1))
            local next_x=$((meteor_x[i] + meteor_dx[i]))
            if (( next_y < HEIGHT && next_x >= 0 && next_x < WIDTH )); then
                new_x+=($next_x); new_y+=($next_y); new_dx+=(${meteor_dx[i]})
            fi
        done
        meteor_x=("${new_x[@]}"); meteor_y=("${new_y[@]}"); meteor_dx=("${new_dx[@]}")
    }

    update_pickups() {
        local new_x=() new_y=()
        for i in "${!pickup_x[@]}"; do
            local next_y=$((pickup_y[i] + 1))
            if (( next_y < HEIGHT )); then
                new_x+=(${pickup_x[i]}); new_y+=($next_y)
            fi
        done
        pickup_x=("${new_x[@]}"); pickup_y=("${new_y[@]}")
    }

    update_bullets() {
        local new_x=() new_y=()
        for i in "${!bullet_x[@]}"; do
            local next_y=$((bullet_y[i] - 1))
            if (( next_y >= 0 )); then
                new_x+=(${bullet_x[i]}); new_y+=($next_y)
            fi
        done
        bullet_x=("${new_x[@]}"); bullet_y=("${new_y[@]}")
    }

    spawn_meteor() {
        meteor_x+=( $((RANDOM % WIDTH)) ); meteor_y+=(0); meteor_dx+=( $((RANDOM % 3 - 1)) )
    }

    spawn_pickup() {
        pickup_x+=( $((RANDOM % WIDTH)) ); pickup_y+=(0)
    }

    spawn_bullet() {
        bullet_x+=($player_pos); bullet_y+=($((HEIGHT - 2)))
    }

    # --- Main Game Loop ---
    stty -icanon -echo
    tput civis
    clear
    echo "Get ready..."
    sleep 1

    while true; do
        tick=$((tick + 1))

        # 1. Handle Input (non-blocking)
        # The `emulate ksh` command at the top of the function ensures that
        # `read -t` works correctly in both Bash and Zsh.
        local key=""
        IFS= read -s -n 1 -t 0.01 key < /dev/tty

        case "$key" in
            a|A) ((player_pos > 0)) && player_pos=$((player_pos - 1));;
            d|D) ((player_pos < WIDTH - 1)) && player_pos=$((player_pos + 1));;
            ' ') spawn_bullet;;
            q|Q) game_over=1;;
        esac

        # 2. Update Object Positions on certain ticks
        if (( tick % meteor_update_rate == 0 )); then update_meteors; fi
        if (( tick % pickup_update_rate == 0 )); then update_pickups; fi
        if (( tick % bullet_update_rate == 0 )); then update_bullets; fi
        
        # 3. Spawn New Objects
        if (( tick % meteor_spawn_rate == 0 )); then spawn_meteor; fi
        if (( tick % pickup_spawn_rate == 0 )); then spawn_pickup; fi
        
        # 4. Handle Collisions and Collections
        if (( ${#bullet_x[@]} > 0 && ${#meteor_x[@]} > 0 )); then
            local meteors_to_remove=()
            local bullets_to_remove=()

            for i in "${!meteor_x[@]}"; do
                for j in "${!bullet_x[@]}"; do
                    local bullet_is_spent=0
                    for spent_j in "${bullets_to_remove[@]}"; do
                        if (( j == spent_j )); then bullet_is_spent=1; break; fi
                    done
                    if (( bullet_is_spent )); then continue; fi

                    if (( meteor_x[i] == bullet_x[j] && meteor_y[i] == bullet_y[j] )); then
                        meteors_to_remove+=($i)
                        bullets_to_remove+=($j)
                        score=$((score + 25))
                        break
                    fi
                done
            done

            if (( ${#meteors_to_remove[@]} > 0 )); then
                local new_mx=(); local new_my=(); local new_mdx=()
                for i in "${!meteor_x[@]}"; do
                    local is_removed=0
                    for remove_idx in "${meteors_to_remove[@]}"; do
                        if (( i == remove_idx )); then is_removed=1; break; fi
                    done
                    if (( !is_removed )); then
                        new_mx+=(${meteor_x[i]}); new_my+=(${meteor_y[i]}); new_mdx+=(${meteor_dx[i]})
                    fi
                done
                meteor_x=("${new_mx[@]}"); meteor_y=("${new_my[@]}"); meteor_dx=("${new_mdx[@]}")
            fi

            if (( ${#bullets_to_remove[@]} > 0 )); then
                local new_bx=(); local new_by=()
                for i in "${!bullet_x[@]}"; do
                    local is_removed=0
                    for remove_idx in "${bullets_to_remove[@]}"; do
                        if (( i == remove_idx )); then is_removed=1; break; fi
                    done
                    if (( !is_removed )); then
                        new_bx+=(${bullet_x[i]}); new_by+=(${bullet_y[i]})
                    fi
                done
                bullet_x=("${new_bx[@]}"); bullet_y=("${new_by[@]}")
            fi
        fi

        local new_px=(); local new_py=()
        for i in "${!pickup_x[@]}"; do
            if (( player_pos == pickup_x[i] && HEIGHT - 1 == pickup_y[i] )); then
                score=$((score + 10))
            else
                new_px+=(${pickup_x[i]}); new_py+=(${pickup_y[i]})
            fi
        done
        pickup_x=("${new_px[@]}"); pickup_y=("${new_py[@]}")

        # 5. Check for Level Up
        if (( score >= score_to_next_level )); then
            level=$((level + 1))
            score_to_next_level=$((score_to_next_level + 100))
            (( meteor_spawn_rate > 4 )) && meteor_spawn_rate=$((meteor_spawn_rate - 1))
            (( meteor_update_rate > 1 )) && meteor_update_rate=$((meteor_update_rate - 1))
        fi

        # 6. Prepare Screen Buffer
        local screen=()
        for ((i=0; i<WIDTH*HEIGHT; i++)); do screen[i]=$EMPTY_CHAR; done
        for i in "${!meteor_x[@]}"; do screen[meteor_y[i]*WIDTH+meteor_x[i]]=$METEOR_CHAR; done
        for i in "${!pickup_x[@]}"; do screen[pickup_y[i]*WIDTH+pickup_x[i]]=$PICKUP_CHAR; done
        for i in "${!bullet_x[@]}"; do screen[bullet_y[i]*WIDTH+bullet_x[i]]=$BULLET_CHAR; done

        # 7. Check for Player <-> Meteor Collision
        local player_index=$(((HEIGHT - 1) * WIDTH + player_pos))
        if [[ "${screen[player_index]}" == "$METEOR_CHAR" ]]; then game_over=1; fi
        screen[$player_index]=$PLAYER_CHAR

        # 8. Draw the Screen
        draw_screen
        
        # 9. Check Game Over
        if (( game_over )); then break; fi

        # 10. Control frame rate
        sleep "$GAME_LOOP_SLEEP"
    done

    # --- Game Over Screen ---
    sleep 0.5
    clear
    echo "============================="
    echo "        GAME OVER"
    echo "============================="
    echo "  Your final score: $score"
    echo "   You reached level $level!"
    echo "============================="
    echo

    # Unset the trap when the function finishes
    trap - EXIT
}
