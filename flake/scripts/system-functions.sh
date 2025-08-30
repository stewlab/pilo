#!/usr/bin/env bash

# A detailed system information script without ASCII art.
# Works on Linux and macOS, providing deeper hardware and software info.

# This function prints a formatted label and value.
print_info() {
    local C_LABEL='\033[1;34m' # Bold Blue
    local C_VALUE='\033[0;37m' # White
    local C_RESET='\033[0m'

    printf "%b%-12s%b%s%b\n" "$C_LABEL" "$1" "$C_VALUE" "$2" "$C_RESET"
}

# --- Main Information Gathering Function ---
sys_info() {
    echo # Start with a blank line

    # --- User & OS Info ---
    print_info "Host:" "$(whoami)@$(hostname)"
    print_info "Uptime:" "$(uptime -p 2>/dev/null || uptime | awk -F'(up |[,])' '{print $2}' | sed 's/^[ \t]*//g')"
    print_info "Kernel:" "$(uname -s) $(uname -r)"
    print_info "Arch:" "$(uname -m)"
    print_info "Shell:" "${SHELL##*/}"

    # --- OS-Specific Section ---
    case "$(uname -s)" in
        Linux)
            # OS / Distro
            if [ -f /etc/os-release ]; then
                . /etc/os-release
                print_info "OS:" "$PRETTY_NAME"
            fi
            
            # Desktop Environment
            if [ -n "$XDG_CURRENT_DESKTOP" ]; then
                print_info "DE:" "$XDG_CURRENT_DESKTOP"
            fi

            # Package Count
            pkg_count=""
            if command -v dpkg &>/dev/null; then
                pkg_count="$(dpkg-query -f '.' -W | wc -l) (dpkg)"
            elif command -v rpm &>/dev/null; then
                pkg_count="$(rpm -qa | wc -l) (rpm)"
            elif command -v pacman &>/dev/null; then
                pkg_count="$(pacman -Qq | wc -l) (pacman)"
            fi
            [ -n "$pkg_count" ] && print_info "Packages:" "$pkg_count"

            # CPU Info
            cpu_model=$(grep 'model name' /proc/cpuinfo | uniq | cut -d ':' -f 2 | sed 's/^[ \t]*//')
            cpu_cores=$(nproc --all 2>/dev/null || grep -c ^processor /proc/cpuinfo)
            print_info "CPU:" "$cpu_model ($cpu_cores)"
            
            # GPU Info
            if command -v lspci &>/dev/null; then
                gpu_info=$(lspci | grep -i 'vga\|3d\|display' | head -n 1 | cut -d ':' -f 3 | sed 's/^[ \t]*//')
                print_info "GPU:" "$gpu_info"
            fi
            ;;

        Darwin)
            # OS
            os_name=$(sw_vers -productName)
            os_version=$(sw_vers -productVersion)
            print_info "OS:" "$os_name $os_version"
            
            # Window Manager (macOS has a standard one)
            print_info "WM:" "Quartz Compositor"
            
            # Package Count (Homebrew)
            if command -v brew &>/dev/null; then
                formulae=$(brew list --formula | wc -l)
                casks=$(brew list --cask | wc -l 2>/dev/null || echo 0)
                print_info "Packages:" "$((formulae + casks)) (brew)"
            fi

            # CPU Info
            cpu_model=$(sysctl -n machdep.cpu.brand_string)
            cpu_cores=$(sysctl -n hw.ncpu)
            print_info "CPU:" "$cpu_model ($cpu_cores)"
            
            # GPU Info
            gpu_info=$(system_profiler SPDisplaysDataType 2>/dev/null | awk -F': ' '/Chipset Model/ {print $2}')
            print_info "GPU:" "$gpu_info"
            ;;
    esac
    
    # --- Hardware Info (Common) ---
    # Memory (MiB) with percentage
    if [[ "$(uname -s)" == "Linux" ]]; then
        mem_total_kib=$(grep MemTotal /proc/meminfo | awk '{print $2}')
        mem_avail_kib=$(grep MemAvailable /proc/meminfo | awk '{print $2}')
        mem_used_mib=$(((mem_total_kib - mem_avail_kib) / 1024))
        mem_total_mib=$((mem_total_kib / 1024))
        if (( mem_total_mib > 0 )); then
            mem_perc=$((mem_used_mib * 100 / mem_total_mib))
            print_info "Memory:" "${mem_used_mib}MiB / ${mem_total_mib}MiB (${mem_perc}%)"
        else
            print_info "Memory:" "${mem_used_mib}MiB / ${mem_total_mib}MiB"
        fi
    else # macOS
        # Parse 'top' command for memory usage on macOS
        mem_info=$(top -l 1 | grep PhysMem | awk '{
            used_val = substr($2, 1, length($2)-1); used_unit = substr($2, length($2));
            unused_val = substr($6, 1, length($6)-1); unused_unit = substr($6, length($6));
            
            used_mib=used_val; if(used_unit=="G") used_mib*=1024; if(used_unit=="K") used_mib/=1024;
            unused_mib=unused_val; if(unused_unit=="G") unused_mib*=1024; if(unused_unit=="K") unused_mib/=1024;

            total_mib = used_mib + unused_mib;
            perc = (used_mib * 100) / total_mib;
            printf "%.0fMiB / %.0fMiB (%.0f%%)", used_mib, total_mib, perc;
        }')
        print_info "Memory:" "$mem_info"
    fi

    # Disk Usage for /
    disk_info=$(df -k / | awk 'NR==2 {printf "%.1fGiB / %.1fGiB (%.0f%%)", $3/1024/1024, ($3+$4)/1024/1024, $5}')
    print_info "Disk (/):" "$disk_info"

    echo # End with a blank line
}

function greet_user() {
  echo "Hello, $USER! Welcome to your NixOS system."
}
