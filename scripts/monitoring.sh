#!/bin/bash
# ==============================================================================
# X / LOSTMEDIA.PRO — PRODUCTION SYSTEM & CONTAINER MONITORING
# ==============================================================================

clear
echo -e "\033[1;36m==============================================================================\033[0m"
echo -e "\033[1;36m  X PLATFORM (https://lostmedia.pro) — LIVE MONITORING DASHBOARD              \033[0m"
echo -e "\033[1;36m==============================================================================\033[0m"

# 1. System Resources
echo -e "\n\033[1;33m[1] VPS SYSTEM RESOURCES\033[0m"
UPTIME=$(uptime -p)
LOAD=$(uptime | awk -F'load average:' '{ print $2 }')
echo "  Uptime       : $UPTIME"
echo "  Load Average :$LOAD"
echo ""
free -h | awk 'NR==1{print "               " $1 "       " $2 "       " $3 "       " $4 "   " $5 "   " $6} NR>1{printf "  %-12s: %-10s %-10s %-10s\n", $1, $2, $3, $4}'
echo ""
df -h / | awk 'NR==1{printf "  %-12s: %-8s %-8s %-8s %-6s\n", "Disk", $2, $3, $4, $5} NR==2{printf "  /dev/root   : %-8s %-8s %-8s %-6s\n", $2, $3, $4, $5}'

# 2. Docker Containers Stats
echo -e "\n\033[1;33m[2] RUNNING X CONTAINERS (RESOURCE CONSUMPTION)\033[0m"
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.PIDs}}"

# 3. Health & Readiness
echo -e "\n\033[1;33m[3] APPLICATION HEALTH CHECK\033[0m"
HEALTH=$(curl -s -k https://127.0.0.1/health/ready)
echo "  Endpoint: https://lostmedia.pro/health/ready"
echo "  Payload : $HEALTH"

# 4. Metrics & Goroutines
echo -e "\n\033[1;33m[4] BACKEND RUNTIME METRICS\033[0m"
METRICS=$(curl -s -k https://127.0.0.1/health/metrics)
echo "  Payload : $METRICS"

# 5. Firewall & Public Ports
echo -e "\n\033[1;33m[5] FIREWALL & LISTENING PORTS (ANTI-EXPOSE)\033[0m"
sudo ufw status | grep -E "Status|22/tcp|80|443" | head -n 5 | awk '{print "  " $0}'
echo ""
echo -e "\033[1;36m==============================================================================\033[0m"
echo -e "Tip: Untuk monitoring streaming real-time secara interaktif, jalankan: \033[1;32mdocker stats\033[0m"
echo -e "\033[1;36m==============================================================================\033[0m"
