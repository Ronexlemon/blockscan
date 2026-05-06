package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)


var (
    // ProcessBlock
    BlocksProcessed = promauto.NewCounter(prometheus.CounterOpts{
        Name: "pipeline_blocks_processed_total",
        Help: "Total blocks processed",
    })
    BlockProcessDuration = promauto.NewHistogram(prometheus.HistogramOpts{
        Name:    "pipeline_block_process_duration_seconds",
        Help:    "Time to fully process one block",
        Buckets: prometheus.DefBuckets,
    })

    // Transactions
    TxTotal = promauto.NewCounter(prometheus.CounterOpts{
        Name: "pipeline_transactions_total",
        Help: "Total transactions seen across all blocks",
    })
    TxSkipped = promauto.NewCounter(prometheus.CounterOpts{
        Name: "pipeline_transactions_skipped_total",
        Help: "Transactions skipped (system addresses / selectors)",
    })

    // Method calls
    MethodCallsFound = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "pipeline_method_calls_total",
        Help: "ERC20 method calls decoded, by method name",
    }, []string{"method"})

    DecodedCallsTotal = promauto.NewCounter(prometheus.CounterOpts{
        Name: "pipeline_decoded_calls_total",
        Help: "Successfully decoded calls",
    })
    DecodeErrors = promauto.NewCounter(prometheus.CounterOpts{
        Name: "pipeline_decode_errors_total",
        Help: "Calls that failed DecodeMethodCall",
    })

    // Event logs
    EventLogsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "pipeline_event_logs_total",
        Help: "Event logs collected, by topic",
    }, []string{"topic"})
    FallbackUsed = promauto.NewCounter(prometheus.CounterOpts{
        Name: "pipeline_event_log_fallback_total",
        Help: "Times FilterLogs failed and per-tx fallback was used",
    })

    // RPC
    RPCCalls = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "pipeline_rpc_calls_total",
        Help: "RPC calls made, by method and status",
    }, []string{"method", "status"})
    RPCDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "pipeline_rpc_duration_seconds",
        Help:    "RPC call latency",
        Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5},
    }, []string{"method"})

    // Empty blocks
    EmptyBlocks = promauto.NewCounter(prometheus.CounterOpts{
        Name: "pipeline_empty_blocks_total",
        Help: "Blocks with no ERC20 activity",
    })
)