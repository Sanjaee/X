# LostMedia Production Incident Response Runbook

**Target**: `https://lostmedia.pro/`  
**Host**: Linux Ubuntu 4GB VPS (`103.89.2.68`)  

---

## 1. Incident Severity Definitions & Response SLAs

| Severity | Definition | Response SLA | Mitigation Action |
| :--- | :--- | :--- | :--- |
| **SEV-1 (Critical)** | Core service down, DB offline, active security breach, data corruption | < 15 minutes | Immediate containment, failover, rollback |
| **SEV-2 (High)** | API error rate > 5%, payment webhook failure, rate limit deadlock | < 30 minutes | Hotfix, traffic throttling, component restart |
| **SEV-3 (Medium)**| Latency spike (p95 > 500ms), single non-critical endpoint erroring | < 2 hours | Resource inspection, query optimization |
| **SEV-4 (Low)** | Non-impacting warning, single client anomaly, minor log noise | Next business day| Log review, monitoring adjustment |

---

## 2. Emergency Operational Playbooks

### Playbook A: Application Error Rate Spike (HTTP 5xx > 1%)
1. **Inspect Live Container Status**:
   ```bash
   ssh zacode@103.89.2.68 "docker ps; docker stats --no-stream"
   ```
2. **Review Recent API Logs**:
   ```bash
   ssh zacode@103.89.2.68 "docker logs --tail 100 x-api"
   ```
3. **Check Database Health**:
   ```bash
   ssh zacode@103.89.2.68 "docker exec x-postgres psql -U postgres -d x_prod -c 'SELECT count(*) FROM pg_stat_activity WHERE state = '\''active\'\'';'"
   ```
4. **Graceful Service Restart** (if memory lock or deadlock detected):
   ```bash
   ssh zacode@103.89.2.68 "docker restart x-api"
   ```

### Playbook B: VPS Memory Saturation / OOM Spike (> 85% RAM)
1. **Identify Highest Memory Consumer**:
   ```bash
   ssh zacode@103.89.2.68 "ps aux --sort=-%mem | head -n 10"
   ```
2. **Check Redis Cache Footprint**:
   ```bash
   ssh zacode@103.89.2.68 "docker exec x-redis redis-cli -a supersecretredispass info memory"
   ```
3. **If Redis Exceeds Quota**: Trigger cache eviction:
   ```bash
   ssh zacode@103.89.2.68 "docker exec x-redis redis-cli -a supersecretredispass memory purge"
   ```

### Playbook C: Rollback Procedure
If a production deployment introduces a regression:
1. Identify the previous working git commit:
   ```bash
   git log -n 5 --oneline
   ```
2. Checkout previous commit:
   ```bash
   git checkout <PREVIOUS_COMMIT_SHA>
   ```
3. Rebuild and restart containers cleanly:
   ```bash
   docker compose -f deploy/docker/docker-compose.prod.yml build
   docker compose -f deploy/docker/docker-compose.prod.yml up -d --force-recreate
   ```
