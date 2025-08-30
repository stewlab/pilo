#!/usr/bin/env bash

# A Tetris-like game written in pure Bash.
# This script defines the `block_puzzle` function. To play, source this file
# and then run the command: block_puzzle

block_puzzle() {
    # --- Shell Compatibility ---
    # If in Zsh, set options to ensure Bash-like array indexing (0-based).
    if [ -n "$ZSH_VERSION" ]; then
        setopt local_options ksh_arrays
    fi

    # --- Game Configuration ---
    local W=10 H=20
    local BORDER_CHAR="■"
    local PIECE_CHAR="■"
    local GAME_LOOP_SLEEP=0.01

    # Tetromino definitions (4×4 blocks in row-major order)
    local tetros=(
        "....#...####................." ".#...###....................."
        "##..##......................" ".##..##....................."
        "##...##....................." "#...###....................."
        "..#..###....................."
    )
    # Colors for each piece (ANSI escape codes)
    local colors=(
        "\033[38;5;39m" "\033[38;5;129m" "\033[38;5;226m" "\033[38;5;196m"
        "\033[38;5;46m" "\033[38;5;208m" "\033[38;5;21m"
    )
    local BORDER_COLOR="\033[38;5;242m"
    local RESET_COLOR="\033[0m"

    # --- Game State ---
    # Shell-aware declaration of the associative array for the board.
    if [ -n "$ZSH_VERSION" ]; then
        typeset -A board
    else
        declare -A board
    fi
    local px py rot piece next_piece
    local score=0 level=1 lines_cleared=0
    local game_over=0 tick=0
    local gravity_rate=20 # Piece falls every N ticks. Lower is faster.

    # --- Terminal Setup & Cleanup ---
    cleanup() {
        stty sane # Restore terminal settings
        tput cnorm # Show cursor
        tput sgr0  # Reset text formatting
        clear
    }
    trap cleanup EXIT

    # --- Core Game Functions ---
    get_cell() {
        local p_id=$1 r=$2 rx=$3 ry=$4 idx
        case $r in
            1) idx=$(((3 - ry) * 4 + rx));;
            2) idx=$((15 - ry * 4 - rx));;
            3) idx=$((ry * 4 + (3 - rx)));;
            *) idx=$((ry * 4 + rx));;
        esac
        printf '%s' "${tetros[$p_id]:idx:1}"
    }

    check_collision() {
        local check_px=$1 check_py=$2 check_rot=$3
        for ry in 0 1 2 3; do
            for rx in 0 1 2 3; do
                if [[ "$(get_cell "$piece" "$check_rot" "$rx" "$ry")" == '#' ]]; then
                    local board_x=$((check_px + rx))
                    local board_y=$((check_py + ry))
                    if (( board_x < 0 || board_x >= W || board_y >= H )) || \
                       [[ -n "${board[$((board_y * W + board_x))]}" ]]; then
                        return 1 # Collision detected
                    fi
                fi
            done
        done
        return 0 # No collision
    }

    spawn_piece() {
        piece=${next_piece:-$((RANDOM % 7))}
        next_piece=$((RANDOM % 7))
        rot=0
        px=$((W / 2 - 2))
        py=0
        if ! check_collision "$px" "$py" "$rot"; then
            game_over=1
        fi
    }

    # --- Drawing Function ---
    draw_screen() {
        local buffer=""
        buffer+="Score: $score   Level: $level   Lines: $lines_cleared\n"
        buffer+="${BORDER_COLOR}${BORDER_CHAR}"
        for ((i=0; i<W*2; i++)); do buffer+="─"; done
        buffer+="${BORDER_CHAR}${RESET_COLOR}   Next:\n"

        for ((y=0; y<H; y++)); do
            buffer+="${BORDER_COLOR}│${RESET_COLOR}"
            for ((x=0; x<W; x++)); do
                if [[ -n "${board[$((y * W + x))]}" ]]; then
                    buffer+="${board[$((y * W + x))]} "
                else
                    buffer+="  "
                fi
            done
            buffer+="${BORDER_COLOR}│${RESET_COLOR}"

            if (( y > 0 && y < 6 )); then
                buffer+="   "
                for ((nx=0; nx<4; nx++)); do
                    if [[ "$(get_cell "$next_piece" 0 "$nx" "$((y-1))")" == '#' ]]; then
                        buffer+="${colors[$next_piece]}$PIECE_CHAR ${RESET_COLOR}"
                    else
                        buffer+="  "
                    fi
                done
            fi
            buffer+="\n"
        done

        buffer+="${BORDER_COLOR}${BORDER_CHAR}"
        for ((i=0; i<W*2; i++)); do buffer+="─"; done
        buffer+="${BORDER_CHAR}${RESET_COLOR}\n"
        buffer+="Controls: a/d/s/w, space=Drop, q=Quit\n"

        local piece_buffer=""
        for ry in 0 1 2 3; do
            for rx in 0 1 2 3; do
                if [[ "$(get_cell "$piece" "$rot" "$rx" "$ry")" == '#' ]]; then
                    piece_buffer+="\033[$((py + ry + 3));$(((px + rx) * 2 + 2))H"
                    piece_buffer+="${colors[$piece]}$PIECE_CHAR ${RESET_COLOR}"
                fi
            done
        done
        
        tput cup 0 0
        printf "%b%b" "$buffer" "$piece_buffer"
    }

    # --- Main Game ---
    stty -icanon -echo; tput civis; clear
    spawn_piece

    while (( !game_over )); do
        tick=$((tick + 1))
        
        # Handle Input (Shell-aware)
        local key=""
        if [ -n "$ZSH_VERSION" ]; then
            zmodload zsh/zselect 2>/dev/null || true
            if zselect -t 1 -r; then
                read -s -k 1 key < /dev/tty
            fi
        else
            read -s -n 1 -t 0.01 key < /dev/tty
        fi

        case "$key" in
            a|A) check_collision "$((px - 1))" "$py" "$rot" && ((px--));;
            d|D) check_collision "$((px + 1))" "$py" "$rot" && ((px++));;
            w|W) 
                local new_rot=$(((rot + 1) % 4))
                check_collision "$px" "$py" "$new_rot" && rot=$new_rot
                ;;
            s|S) 
                if check_collision "$px" "$((py + 1))" "$rot"; then
                    ((py++))
                    tick=0 
                fi
                ;;
            ' ') 
                while check_collision "$px" "$((py + 1))" "$rot"; do
                    ((py++))
                done
                tick=$((gravity_rate - 1))
                ;;
            q|Q) break;;
        esac

        # Gravity based on tick
        if (( tick % gravity_rate == 0 )); then
            if check_collision "$px" "$((py + 1))" "$rot"; then
                ((py++))
            else
                for ry in 0 1 2 3; do
                    for rx in 0 1 2 3; do
                        if [[ "$(get_cell "$piece" "$rot" "$rx" "$ry")" == '#' ]]; then
                            board[$(((py + ry) * W + px + rx))]="${colors[$piece]}$PIECE_CHAR"
                        fi
                    done
                done

                local lines_cleared_this_turn=0
                for ((y=0; y<H; y++)); do
                    local full=1
                    for ((x=0; x<W; x++)); do
                        [[ -z "${board[$((y * W + x))]}" ]] && { full=0; break; }
                    done
                    if (( full )); then
                        ((lines_cleared_this_turn++))
                        for ((yy=y; yy>0; yy--)); do
                            for ((x=0; x<W; x++)); do
                                board[$((yy * W + x))]=${board[$(((yy - 1) * W + x))]}
                            done
                        done
                        for ((x=0; x<W; x++)); do unset board[$x]; done
                    fi
                done

                if (( lines_cleared_this_turn > 0 )); then
                    lines_cleared=$((lines_cleared + lines_cleared_this_turn))
                    score=$((score + (lines_cleared_this_turn * lines_cleared_this_turn * 10)))
                    if (( lines_cleared >= level * 10 )); then
                        ((level++))
                        (( gravity_rate > 5 )) && gravity_rate=$((gravity_rate - 2))
                    fi
                fi
                spawn_piece
            fi
        fi
        
        draw_screen
        sleep "$GAME_LOOP_SLEEP"
    done

    # --- Game Over Screen ---
    # The trap will call cleanup, which clears the screen.
    # To show the message, we must restore the terminal first.
    stty sane; tput cnorm; tput sgr0; clear
    echo "============================="
    echo "         GAME OVER"
    echo "============================="
    echo "  Final Score: $score"
    echo "  Level: $level"
    echo "  Lines Cleared: $lines_cleared"
    echo "============================="
}

# To play, source this file and run `block_puzzle`
# block_puzzle
