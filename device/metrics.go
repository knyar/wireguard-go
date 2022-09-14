package device

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	peerBytesReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wireguard_peer_received_bytes_total",
		Help: "Total number of received bytes",
	}, []string{"peer", "endpoint"})
	peerBytesDropped = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wireguard_peer_deduplicated_bytes_total",
		Help: "Total number of received bytes dropped by deduplication",
	}, []string{"peer", "endpoint"})
	peerBytesSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wireguard_peer_sent_bytes_total",
		Help: "Total number of received bytes",
	}, []string{"peer"})
	peerPacketsReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wireguard_peer_received_packets_total",
		Help: "Total number of received packets",
	}, []string{"peer", "endpoint"})
	peerPacketsDropped = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wireguard_peer_deduplicated_packets_total",
		Help: "Total number of received packets dropped by deduplication",
	}, []string{"peer", "endpoint"})
	peerPacketsSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wireguard_peer_sent_packets_total",
		Help: "Total number of received packets",
	}, []string{"peer"})
)
