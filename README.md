# postgresql configure (fork from pgtune)

tuning PostgreSQL config by your hardware

> This is a copy and language rewrite, source code at https://github.com/le0pard/pgtune/blob/master/webpack/selectors/configuration.js

```bash
$ go run main.go -totalMemory=2 -totalMemoryUnit=GB -cpuNum=8
# DB Version: 10
# OS Type: linux
# DB Type: web
# Total Memory (RAM): 2 GB
# CPUs num: 8
# Data Storage: ssd

max_connections = 200
shared_buffers = 512MB
effective_cache_size = 1536MB
maintenance_work_mem = 128MB
checkpoint_completion_target = 0.7
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 655kB
min_wal_size = 1GB
max_wal_size = 2GB
max_worker_processes = 8
max_parallel_workers_per_gather = 4
max_parallel_workers = 8
```

## Usage

```bash
$ ./pgtune-go -h
Usage of ./pgtune-go:
  -connectionNum int
    	Maximum number of PostgreSQL client connections
  -cpuNum int
    	Number of CPUs, which PostgreSQL can use
    	CPUs = threads per core * cores per socket * sockets
  -dbType string
    	For what type of application is PostgreSQL used (default "web")
  -dbVersion string
    	PostgreSQL version (find out via 'SELECT version();') (default "10")
  -hdType string
    	Type of data storage device (default "ssd")
  -osType string
    	Operation system of the PostgreSQL server host (default "linux")
  -totalMemory int
    	How much memory can PostgreSQL use
  -totalMemoryUnit string
    	memory unit (default "GB")
```

## LICENSE

MIT License
