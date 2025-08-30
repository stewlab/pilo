#!/usr/bin/env bash

# Breakout clone in pure Bash.

# --- Common Setup ---
init_terminal() {
    stty -icanon -echo
    tput civis
    clear
}
cleanup() {
    stty sane
    tput cnorm
    tput sgr0
    clear
    echo
}
trap cleanup EXIT

# --- Breakout Clone ---
break_blocks() {
    init_terminal
    local W=40 H=20
    local PADDLE_LEN=7
    local px=$(( (W-PADDLE_LEN)/2 )) py=$((H-1))
    local bx=$((W/2)) by=$((H/2))
    local dx=1 dy=-1
    # bricks: 5 rows, full width
    local bricks=()
    for ((y=1; y<=5; y++)); do
      for ((x=1; x< W-1; x++)); do bricks[y*W+x]=1; done
    done

    while true; do
        # input
        read -rsn1 -t 0.05 key
        [[ $key == q ]] && break
        [[ $key == a ]] && (( px>0 )) && (( px-- ))
        [[ $key == d ]] && (( px < W-PADDLE_LEN )) && (( px++ ))
        # move ball
        bx=$(( bx+dx ))
        by=$(( by+dy ))
        # collisions
        (( bx<=0 || bx>=W-1 )) && dx=$(( -dx ))
        (( by<=0 )) && dy=$(( -dy ))
        # paddle
        if (( by==py-1 && bx>=px && bx<px+PADDLE_LEN )); then dy=-dy; fi
        # bottom death
        (( by>=H )) && break
        # brick collision
        local idx=$((by*W+bx))
        if (( ${bricks[idx]:-0} == 1 )); then
            bricks[idx]=0; dy=$(( -dy )); fi
        # draw
        tput cup 0 0
        # top border
        printf '%*s
' $((W+2)) '' | tr ' ' '─'
        for ((y=0; y< H; y++)); do
            printf '│'
            for ((x=0; x< W; x++)); do
                idx=$((y*W+x))
                if (( y==by && x==bx )); then printf 'o'
                elif (( y==py && x>=px && x<px+PADDLE_LEN )); then printf '='
                elif (( ${bricks[idx]:-0} )); then printf '█'
                else printf ' '
                fi
            done
            printf '│\n'
        done
        printf '%*s
' $((W+2)) '' | tr ' ' '─'
    done
    sleep 0.5
    echo 'Game Over: Break Blocks'
}

