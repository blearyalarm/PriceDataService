# Price Tracking Service
This is a Golang microservice that serves 2 GRPC APIs:
```
service PriceDataService {
  rpc FindData(FindDataRequest) returns(FindDataResponse);

  // webhook api for scheduler to load data
  rpc LoadData(LoadDataRequest) returns(LoadDataResponse);
}
```

## Setup Dev Environment
Follow the instruction in the parent [readme file](../README.md#setup-dev-environment).

### Start Server
1. Generate protobuf stub files
```shell
# go back to monorepo root folder
cd .. 
make

# go back to zeonology folder
cd zeonology
```

2. Download dependencies
```shell
make tidy
```

3. Start docker containers for local development
```shell
make local
```

4. Migrate database up
```shell
make migrate_up
```

### How to add tracing?
The [&Name] Go Service are layered. It contains: Handler, Service, Integration layer(DB, Gateway).
For each layer's entry method, you should always add following 2 lines at the beginning of the method.
```go
tracer:= otel.Tracer(cfg.GetTracerName())
ctx, span := tracer.Start(ctx, "UserService.Login")
defer span.End()
```


### How to add log entry?
There are 2 sets of logger methods:
- With GRPC context. With context, the output contains: tracing, grpc method info etc.    
  ```go
   func DebugCtx(ctx context.Context, msg string, fields ...zap.Field)
   func InfoCtx(ctx context.Context, msg string, fields ...zap.Field)
   func WarnCtx(ctx context.Context, msg string, fields ...zap.Field) 
   func ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) 
  ```
- Without context  
  ```go
   func Debug(msg string, fields ...zap.Field)
   func Info(msg string, fields ...zap.Field)
   func Warn(msg string, fields ...zap.Field) 
   func Error(msg string, fields ...zap.Field) 
  ```
Example:
```go
logger.InfoCtx(ctx, "test debug", zap.String("name", "entryName"), zap.Int("type", 1))
logger.Info("test debug2 without context", zap.String("name", "entryName"), zap.Int("type", 1))
logger.InfoCtx(ctx, "test debug3", zap.String("name", "entryName"), zap.Int("type", 1))
logger.Info("test debug3", zap.String("name", "entryName"), zap.Int("type", 1))
logger.Debug("test debug3", zap.String("name", "entryName"), zap.Int("type", 1))
logger.DebugCtx(ctx, "test debug3", zap.String("name", "entryName"), zap.Int("type", 1))
```
Output:
```shell
{"LEVEL":"INFO","TIME":"2021-12-27T14:31:09.556-0500","CALLER":"user/user.go:128","MESSAGE":"test debug","grpc.start_time":"2021-12-27T14:31:09-05:00","system":"grpc","span.kind":"server","grpc.service":"usersvc.usersvc","grpc.method":"AuthnWithPassword","peer.address":"[::1]:62214","trace.traceid":"65fca735a96a2980","trace.spanid":"65fca735a96a2980","trace.sampled":"true","name":"entryName","type":1}
{"LEVEL":"INFO","TIME":"2021-12-27T14:31:09.556-0500","CALLER":"user/user.go:129","MESSAGE":"test debug2 without context","name":"entryName","type":1}
{"LEVEL":"INFO","TIME":"2021-12-27T14:31:09.556-0500","CALLER":"user/user.go:130","MESSAGE":"test debug3","grpc.start_time":"2021-12-27T14:31:09-05:00","system":"grpc","span.kind":"server","grpc.service":"usersvc.usersvc","grpc.method":"AuthnWithPassword","peer.address":"[::1]:62214","trace.traceid":"65fca735a96a2980","trace.spanid":"65fca735a96a2980","trace.sampled":"true","name":"entryName","type":1}
{"LEVEL":"INFO","TIME":"2021-12-27T14:31:09.556-0500","CALLER":"user/user.go:131","MESSAGE":"test debug3","name":"entryName","type":1}
{"LEVEL":"DEBUG","TIME":"2021-12-27T14:31:09.557-0500","CALLER":"user/user.go:132","MESSAGE":"test debug3","name":"entryName","type":1}
{"LEVEL":"DEBUG","TIME":"2021-12-27T14:31:09.557-0500","CALLER":"user/user.go:133","MESSAGE":"test debug3","grpc.start_time":"2021-12-27T14:31:09-05:00","system":"grpc","span.kind":"server","grpc.service":"usersvc.usersvc","grpc.method":"AuthnWithPassword","trace.sampled":"true","peer.address":"[::1]:62214","trace.traceid":"65fca735a96a2980","trace.spanid":"65fca735a96a2980","name":"entryName","type":1}
```

### How to add metrics?
1. You should add metrics at package level
2. At the package add a go file "metrics.go".
3. Define metrics collector with prefix "[&Name]_".
4. Register collector in the init() method.
Example:  
```go
var (
	UserAuthnCollector = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: metric.METRICS_PREFIX + "user_authn_with_password_count",
		Help: "Total number of user authentication with password.",
	}, []string{"name"})
)

func init() {
	prometheus.DefaultRegisterer.MustRegister(UserAuthnCollector)
}
```
5. Call collect in your code.  
```go
UserAuthnCollector.WithLabelValues("login").Inc()
```

