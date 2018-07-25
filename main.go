package main

import "flag"
import "fmt"
import "strconv"

// postgresql versions
const DEFAULT_DB_VERSION = "10"
var DB_VERSIONS = []string{DEFAULT_DB_VERSION, "9.6", "9.5", "9.4", "9.3", "9.2"}
// os types
const OS_LINUX = "linux"
const OS_WINDOWS = "windows"
// db types
const DB_TYPE_WEB = "web"
const DB_TYPE_OLTP = "oltp"
const DB_TYPE_DW = "dw"
const DB_TYPE_DESKTOP = "desktop"
const DB_TYPE_MIXED = "mixed"
// size units
const SIZE_UNIT_MB = "MB"
const SIZE_UNIT_GB = "GB"
// harddrive types
const HARD_DRIVE_SSD = "ssd"
const HARD_DRIVE_SAN = "san"
const HARD_DRIVE_HDD = "hdd"

func byteSize(size int) string {
	var result int
	if (size % 1024 != 0 || size < 1024) {
		return fmt.Sprintf("%dkB", size)
	}
	result = size / 1024
	if (result % 1024 != 0 || result < 1024) {
		return fmt.Sprintf("%dMB", result)
	}
	result = result / 1024
	if (result % 1024 != 0 || result < 1024) {
		return fmt.Sprintf("%dGB", result)
	}

	return fmt.Sprintf("%d", size)
}

func main() {
	DBVersion := flag.String("dbVersion", DEFAULT_DB_VERSION, "PostgreSQL version (find out via 'SELECT version();')")
	OSType := flag.String("osType", OS_LINUX, "Operation system of the PostgreSQL server host")
	DBType := flag.String("dbType", DB_TYPE_WEB, "For what type of application is PostgreSQL used")
	TotalMemory := flag.Int("totalMemory", 0, "How much memory can PostgreSQL use")
	TotalMemoryUnit := flag.String("totalMemoryUnit", SIZE_UNIT_GB, "memory unit")
	CPUNum := flag.Int("cpuNum", 0, "Number of CPUs, which PostgreSQL can use\nCPUs = threads per core * cores per socket * sockets")
	ConnectionNum := flag.Int("connectionNum", 0, "Maximum number of PostgreSQL client connections")
	HDType := flag.String("hdType", HARD_DRIVE_SSD, "Type of data storage device")

	flag.Parse()

	fmt.Println("# DB Version:", *DBVersion)
	fmt.Println("# OS Type:", *OSType)
	fmt.Println("# DB Type:", *DBType)
	fmt.Println("# Total Memory (RAM):", *TotalMemory, *TotalMemoryUnit)
	fmt.Println("# CPUs num:", *CPUNum)
	if (*ConnectionNum > 0) {
		fmt.Println("# Connection num:", *ConnectionNum)
	}
	fmt.Println("# Data Storage:", *HDType)
	fmt.Println("")

	SIZE_UNIT_MAP := map[string]int{
		"KB": 1024,
		"MB": 1048576,
		"GB": 1073741824,
		"TB": 1099511627776,
	}

	var FinalConnectionNum int
	if (*ConnectionNum < 1) {
		CONNECTION_NUM_MAP := map[string]int{
			DB_TYPE_WEB: 200,
			DB_TYPE_OLTP: 300,
			DB_TYPE_DW: 20,
			DB_TYPE_DESKTOP: 10,
			DB_TYPE_MIXED: 100,
		}
		FinalConnectionNum = CONNECTION_NUM_MAP[*DBType]
	} else {
		FinalConnectionNum = *ConnectionNum
	}
	fmt.Println("max_connections", "=", FinalConnectionNum)

	totalMemoryInBytes := *TotalMemory * SIZE_UNIT_MAP[*TotalMemoryUnit]
	totalMemoryInKb := totalMemoryInBytes / SIZE_UNIT_MAP["KB"]

	var sharedBuffers int
	SHARED_BUFFERS_VALUE_MAP := map[string]int{
		DB_TYPE_WEB: totalMemoryInKb / 4,
		DB_TYPE_OLTP: totalMemoryInKb / 4,
		DB_TYPE_DW: totalMemoryInKb / 4,
		DB_TYPE_DESKTOP: totalMemoryInKb / 16,
		DB_TYPE_MIXED: totalMemoryInKb / 4,
	}
	sharedBuffers = SHARED_BUFFERS_VALUE_MAP[*DBType]
	// Limit shared_buffers to 512MB on Windows
	winMemoryLimit := 512 * SIZE_UNIT_MAP["MB"] / SIZE_UNIT_MAP["KB"]
	if (OS_WINDOWS == *OSType && sharedBuffers > winMemoryLimit) {
		sharedBuffers = winMemoryLimit
	}
	fmt.Println("shared_buffers", "=", byteSize(sharedBuffers))

	var effectiveCacheSize int
	EFFECTIVE_CACHE_SIZE_MAP := map[string]int{
		DB_TYPE_WEB: totalMemoryInKb * 3 / 4,
		DB_TYPE_OLTP: totalMemoryInKb * 3 / 4,
		DB_TYPE_DW: totalMemoryInKb * 3 / 4,
		DB_TYPE_DESKTOP: totalMemoryInKb / 4,
		DB_TYPE_MIXED: totalMemoryInKb * 3 / 4,
	}
	effectiveCacheSize = EFFECTIVE_CACHE_SIZE_MAP[*DBType]
	fmt.Println("effective_cache_size", "=", byteSize(effectiveCacheSize))

	var maintenanceWorkMem int
	MAINTENANCE_WORK_MEM_MAP := map[string]int{
		DB_TYPE_WEB: totalMemoryInKb / 16,
		DB_TYPE_OLTP: totalMemoryInKb / 16,
		DB_TYPE_DW: totalMemoryInKb / 8,
		DB_TYPE_DESKTOP: totalMemoryInKb/ 16,
		DB_TYPE_MIXED: totalMemoryInKb / 16,
	}
	maintenanceWorkMem = MAINTENANCE_WORK_MEM_MAP[*DBType]
	// Cap maintenance RAM at 2GB on servers with lots of memory
	memoryLimit := 2 * SIZE_UNIT_MAP["GB"] / SIZE_UNIT_MAP["KB"]
	if (maintenanceWorkMem > memoryLimit) {
		maintenanceWorkMem = memoryLimit
	}
	fmt.Println("maintenance_work_mem", "=", byteSize(maintenanceWorkMem))

	DBVersionFloat, _ := strconv.ParseFloat(*DBVersion, 32)

	var checkpointCompletionTarget float32
	CHECKPOINT_COMPLETION_TARGET_MAP := map[string]float32{
		DB_TYPE_WEB: 0.7,
		DB_TYPE_OLTP: 0.9,
		DB_TYPE_DW: 0.9,
		DB_TYPE_DESKTOP: 0.5,
		DB_TYPE_MIXED: 0.9,
	}
	checkpointCompletionTarget = CHECKPOINT_COMPLETION_TARGET_MAP[*DBType]
	fmt.Println("checkpoint_completion_target", "=", checkpointCompletionTarget)

	var walBuffersValue int
	// Follow auto-tuning guideline for wal_buffers added in 9.1, where it's
	// set to 3% of shared_buffers up to a maximum of 16MB.
	walBuffersValue = 3 * sharedBuffers / 100
	maxWalBuffer := 16 * SIZE_UNIT_MAP["MB"] / SIZE_UNIT_MAP["KB"]
	if (walBuffersValue > maxWalBuffer) {
		walBuffersValue = maxWalBuffer
	}
	// It's nice of wal_buffers is an even 16MB if it's near that number. Since
	// that is a common case on Windows, where shared_buffers is clipped to 512MB,
	// round upwards in that situation
	walBufferNearValue := 14 * SIZE_UNIT_MAP["MB"] / SIZE_UNIT_MAP["KB"]
	if (walBuffersValue > walBufferNearValue && walBuffersValue < maxWalBuffer) {
		walBuffersValue = maxWalBuffer
	}
	// if less, than 32 kb, than set it to minimum
	if (walBuffersValue < 32) {
		walBuffersValue = 32
	}
	fmt.Println("wal_buffers", "=", byteSize(walBuffersValue))

	var defaultStatisticsTarget int
	DEFAULT_STATISTICS_TARGET_MAP := map[string]int{
		DB_TYPE_WEB: 100,
		DB_TYPE_OLTP: 100,
		DB_TYPE_DW: 500,
		DB_TYPE_DESKTOP: 100,
		DB_TYPE_MIXED: 100,
	}
	defaultStatisticsTarget = DEFAULT_STATISTICS_TARGET_MAP[*DBType]
	fmt.Println("default_statistics_target", "=", defaultStatisticsTarget)

	var randomPageCost float32
	RANDOM_PAGE_COST_MAP := map[string]float32 {
		HARD_DRIVE_HDD: 4,
		HARD_DRIVE_SSD: 1.1,
		HARD_DRIVE_SAN: 1.1,
	}
	randomPageCost = RANDOM_PAGE_COST_MAP[*HDType]
	fmt.Println("random_page_cost", "=", randomPageCost)

	var effectiveIoConcurrency float32
	EFFECTIVE_IO_CONCURRENCY := map[string]float32 {
		HARD_DRIVE_HDD: 2,
		HARD_DRIVE_SSD: 200,
		HARD_DRIVE_SAN: 300,
	}

	if (OS_WINDOWS != *OSType) {
		effectiveIoConcurrency = EFFECTIVE_IO_CONCURRENCY[*HDType]
		fmt.Println("effective_io_concurrency", "=", effectiveIoConcurrency)
	}

	var workMemValue int
	var workMemResult int
	var workMemBase int
	if (DBVersionFloat >= 9.5 && *CPUNum >= 2) {
		workMemBase = *CPUNum / 2
	} else {
		workMemBase = 1
	}
	// work_mem is assigned any time a query calls for a sort, or a hash, or any other structure that needs a space allocation, which can happen multiple times per query. So you're better off assuming max_connections * 2 or max_connections * 3 is the amount of RAM that will actually use in reality. At the very least, you need to subtract shared_buffers from the amount you're distributing to connections in work_mem.
	// The other thing to consider is that there's no reason to run on the edge of available memory. If you do that, there's a very high risk the out-of-memory killer will come along and start killing PostgreSQL backends. Always leave a buffer of some kind in case of spikes in memory usage. So your maximum amount of memory available in work_mem should be ((RAM - shared_buffers) / (max_connections * 3) / max_parallel_workers_per_gather).
	workMemValue = (totalMemoryInKb - sharedBuffers) / (FinalConnectionNum * 3) / workMemBase
	
	WORK_MEM_MAP := map[string]int {
		DB_TYPE_WEB: workMemValue,
		DB_TYPE_OLTP: workMemValue,
		DB_TYPE_DW: workMemValue / 2,
		DB_TYPE_DESKTOP: workMemValue / 6,
		DB_TYPE_MIXED: workMemValue / 2,
	}
	workMemResult = WORK_MEM_MAP[*DBType]
	// if less, than 64 kb, than set it to minimum
	if (workMemResult < 64) {
		workMemResult = 64
	}
	fmt.Println("work_mem", "=", byteSize(workMemResult))

	if (DBVersionFloat < 9.5) {
		CHECKPOINT_SEGMENTS_MAP := map[string]int{
			DB_TYPE_WEB: 32,
			DB_TYPE_OLTP: 64,
			DB_TYPE_DW: 128,
			DB_TYPE_DESKTOP: 3,
			DB_TYPE_MIXED: 32,
		}
		fmt.Println("checkpoint_segments", "=", CHECKPOINT_SEGMENTS_MAP[*DBType])
	} else {
		MIN_WAL_SIZE_MAP := map[string]int{
			DB_TYPE_WEB: (1024 * SIZE_UNIT_MAP["MB"] / SIZE_UNIT_MAP["KB"]),
			DB_TYPE_OLTP: (2048 * SIZE_UNIT_MAP["MB"] / SIZE_UNIT_MAP["KB"]),
			DB_TYPE_DW: (4096 * SIZE_UNIT_MAP["MB"] / SIZE_UNIT_MAP["KB"]),
			DB_TYPE_DESKTOP: (100 * SIZE_UNIT_MAP["MB"] / SIZE_UNIT_MAP["KB"]),
			DB_TYPE_MIXED: (1024 * SIZE_UNIT_MAP["MB"] / SIZE_UNIT_MAP["KB"]),
		}
		fmt.Println("min_wal_size", "=", byteSize(MIN_WAL_SIZE_MAP[*DBType]))

		MAX_WAL_SIZE_MAP := map[string]int{
			DB_TYPE_WEB: (2048 * SIZE_UNIT_MAP["MB"] / SIZE_UNIT_MAP["KB"]),
			DB_TYPE_OLTP: (4096 * SIZE_UNIT_MAP["MB"] / SIZE_UNIT_MAP["KB"]),
			DB_TYPE_DW: (8192 * SIZE_UNIT_MAP["MB"] / SIZE_UNIT_MAP["KB"]),
			DB_TYPE_DESKTOP: (1024 * SIZE_UNIT_MAP["MB"] / SIZE_UNIT_MAP["KB"]),
			DB_TYPE_MIXED: (2048 * SIZE_UNIT_MAP["MB"] / SIZE_UNIT_MAP["KB"]),
		}
		fmt.Println("max_wal_size", "=", byteSize(MAX_WAL_SIZE_MAP[*DBType]))
	}

	if (DBVersionFloat >= 9.5 && *CPUNum >= 2) {
		fmt.Println("max_worker_processes", "=", *CPUNum)
  
		if (DBVersionFloat >= 9.6) {
			fmt.Println("max_parallel_workers_per_gather", "=", *CPUNum / 2)
		}
  
		if (DBVersionFloat >= 10) {
			fmt.Println("max_parallel_workers", "=", *CPUNum)
		}
	}

}