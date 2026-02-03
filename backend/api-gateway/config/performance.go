# Performance Optimization Configuration

## Backend Performance Tuning

### Database Connection Pool
```go
# In db/connection.go
sqlDB.SetMaxOpenConns(100)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(time.Hour)
```

### GORM Performance
```go
# Skip default transactions for read operations
db.Session(&gorm.Session{SkipDefaultTransaction: true})
```

### HTTP Server Tuning
```go
# In server/server.go
httpServer := &http.Server{
    Addr:         addr,
    Handler:      h,
    ReadTimeout:  cfg.Server.ReadTimeout,
    WriteTimeout: cfg.Server.WriteTimeout,
    IdleTimeout:  60 * time.Second,
    // Connection limits
    MaxHeaderBytes: 1 << 20, // 1MB
}
```

### Redis Pooling
```go
# In redis/client.go
redisClient := redis.NewClient(&redis.Options{
    PoolSize:     50,
    MinIdleConns:  10,
    MaxRetries:    3,
    DialTimeout:   5 * time.Second,
    ReadTimeout:   3 * time.Second,
    WriteTimeout:  3 * time.Second,
    PoolTimeout:   4 * time.Second,
})
```

## Frontend Performance Tuning

### Build Optimization
# In vite.config.ts
export default defineConfig({
  build: {
    // Code splitting
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react-router-dom'],
          'vendor-antd': ['antd', '@ant-design/icons'],
          'vendor-query': ['@tanstack/react-query'],
        },
      },
    },
    // Chunk size warnings
    chunkSizeWarningLimit: 1000,
    // Minification
    minify: 'esbuild',
    // Source maps for production
    sourcemap: false,
  },
})
```

### React Query Optimization
```typescript
// In api/client.ts
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,      // 5 minutes
      gcTime: 10 * 60 * 1000,         // 10 minutes
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})
```

### Bundle Analysis
```bash
# Analyze bundle size
npm run build
npx vite-bundle-visualizer
```

## Monitoring & Profiling

### CPU Profiling
```go
# Add to main.go
import (
    _ "net/http/pprof"
    _ "net/http/pprof"
)

# Then access:
# curl http://localhost:8080/debug/pprof/
```

### Memory Profiling
```bash
# Generate heap profile
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

### Trace Profiling
```bash
# Start tracing
curl http://localhost:8080/debug/pprof/trace?seconds=30 > trace.out
go tool trace trace.out
```

## Performance Targets

| Metric | Target | Current |
|--------|--------|---------|
| API Response Time (p95) | < 200ms | TBD |
| API Response Time (p99) | < 500ms | TBD |
| Database Query Time | < 50ms | TBD |
| Frontend TTI | < 3s | ~2s |
| Bundle Size | < 2MB | ~1.2MB |
| Memory Usage | < 512MB | TBD |

## Load Testing

### Apache Bench
```bash
# Test API endpoints
ab -n 1000 -c 10 -H "Authorization: Bearer TOKEN" \
   http://localhost:8080/api/v1/hosts
```

### Vegeta
```bash
# Create load test
echo "GET http://localhost:8080/api/v1/clusters" | \
  vegeta attack -duration=30s -rate=50 | \
  vegeta report -type=text
```

## Optimization Checklist

- [ ] Enable database connection pooling
- [ ] Implement Redis caching for frequently accessed data
- [ ] Add HTTP response compression (gzip)
- [ ] Implement pagination for all list endpoints
- [ ] Add database indexes for frequently queried fields
- [ ] Use prepared statements for all database queries
- [ ] Implement rate limiting
- [ ] Add CDN for static assets
- [ ] Enable HTTP/2
- [ ] Implement graceful shutdown
- [ ] Add circuit breakers for external service calls
- [ ] Optimize frontend bundle size
- [ ] Implement lazy loading for routes
- [ ] Add image optimization
- [ ] Implement request debouncing/throttling
