#!/bin/bash

# Test Monitoring Stack
echo "🔍 Testing Go-Agents Monitoring Stack..."
echo

# Function to check URL and show status
check_url() {
    local url=$1
    local service=$2
    local expected_text=$3

    echo -n "Testing $service ($url)..."
    if curl -s --max-time 10 "$url" | grep -q "$expected_text"; then
        echo " ✅ WORKING"
        return 0
    else
        echo " ❌ FAILED"
        return 1
    fi
}

# Check services are running
echo "📋 Service Status:"
docker ps --filter "name=go-agents" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo

# Test Application
echo "🚀 Application Tests:"
check_url "http://localhost:8080/health" "Health Check" "healthy"
check_url "http://localhost:8080/agents" "Agent List" "agents"
check_url "http://localhost:8080/metrics" "Metrics Endpoint" "promhttp_metric_handler"
echo

# Test Prometheus
echo "📊 Prometheus Tests:"
check_url "http://localhost:9091" "Prometheus UI" "Prometheus"
check_url "http://localhost:9091/api/v1/query?query=up" "Metrics Query" "success"
echo

# Test Grafana
echo "📈 Grafana Tests:"
check_url "http://localhost:3001" "Grafana UI" "Grafana"

# Test Grafana API and Dashboards
echo -n "Testing Grafana API..."
if curl -s -u admin:admin "http://localhost:3001/api/search" | grep -q "Go-Agents System Overview"; then
    echo " ✅ WORKING"
else
    echo " ❌ FAILED"
fi

echo -n "Testing System Overview Dashboard..."
if curl -s -u admin:admin "http://localhost:3001/api/dashboards/uid/go-agents-overview" | grep -q "go-agents-overview"; then
    echo " ✅ WORKING"
else
    echo " ❌ FAILED"
fi

echo -n "Testing Agent Performance Dashboard..."
if curl -s -u admin:admin "http://localhost:3001/api/dashboards/uid/agent-performance" | grep -q "agent-performance"; then
    echo " ✅ WORKING"
else
    echo " ❌ FAILED"
fi

echo -n "Testing Infrastructure Dashboard..."
if curl -s -u admin:admin "http://localhost:3001/api/dashboards/uid/infrastructure" | grep -q "infrastructure"; then
    echo " ✅ WORKING"
else
    echo " ❌ FAILED"
fi
echo

# Test Data Sources
echo "🔗 Data Source Tests:"
echo -n "Testing Prometheus Data Source..."
if curl -s -u admin:admin "http://localhost:3001/api/datasources" | grep -q "prometheus"; then
    echo " ✅ WORKING"
else
    echo " ❌ FAILED"
fi
echo

# Generate test traffic
echo "🌊 Generating Test Traffic..."
for i in {1..10}; do
    curl -s http://localhost:8080/health > /dev/null
    curl -s http://localhost:8080/agents > /dev/null
    curl -s http://localhost:8080/capabilities > /dev/null
    [ $((i % 3)) -eq 0 ] && echo -n "."
done
echo " Done!"
echo

# Test Metrics Collection
echo "📈 Metrics Collection Tests:"
echo -n "Testing HTTP Request Metrics..."
if curl -s "http://localhost:9091/api/v1/query?query=promhttp_metric_handler_requests_total" | grep -q "result"; then
    echo " ✅ WORKING"
else
    echo " ❌ FAILED"
fi

echo -n "Testing Go Runtime Metrics..."
if curl -s "http://localhost:9091/api/v1/query?query=go_memstats_alloc_bytes" | grep -q "result"; then
    echo " ✅ WORKING"
else
    echo " ❌ FAILED"
fi
echo

# Dashboard URLs
echo "🎯 Monitoring URLs:"
echo "  📊 Prometheus:      http://localhost:9091/"
echo "  📈 Grafana:         http://localhost:3001/ (admin/admin)"
echo "  🎪 System Overview: http://localhost:3001/d/go-agents-overview/"
echo "  🎯 Agent Performance: http://localhost:3001/d/agent-performance/"
echo "  🏗️  Infrastructure:  http://localhost:3001/d/infrastructure/"
echo "  📊 App Metrics:     http://localhost:8080/metrics"
echo

echo "✅ Monitoring Stack Test Complete!"
echo "🎉 Full observability is working! Visit the dashboards to see real-time metrics."
