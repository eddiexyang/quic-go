package bbr

import (
	"testing"
	"time"


	"github.com/eddiexyang/quic-go/congestion"
)

func TestSetMaxDatagramSizeRescalesPacketSizedWindows(t *testing.T) {
	const oldMaxDatagramSize = congestion.ByteCount(1000)
	const newMaxDatagramSize = congestion.ByteCount(1400)
	const initialCongestionWindowPackets = congestion.ByteCount(20)
	const maxCongestionWindowPackets = congestion.ByteCount(80)

	b := newBbrSender(
		DefaultClock{},
		oldMaxDatagramSize,
		initialCongestionWindowPackets*oldMaxDatagramSize,
		maxCongestionWindowPackets*oldMaxDatagramSize,
		ProfileStandard,
	)
	b.congestionWindow = b.initialCongestionWindow

	b.SetMaxDatagramSize(newMaxDatagramSize)

	if (initialCongestionWindowPackets*newMaxDatagramSize) != (b.initialCongestionWindow) { t.Fatalf("got %v, want %v", b.initialCongestionWindow, initialCongestionWindowPackets*newMaxDatagramSize) }
	if (maxCongestionWindowPackets*newMaxDatagramSize) != (b.maxCongestionWindow) { t.Fatalf("got %v, want %v", b.maxCongestionWindow, maxCongestionWindowPackets*newMaxDatagramSize) }
	if (minCongestionWindowPackets*newMaxDatagramSize) != (b.minCongestionWindow) { t.Fatalf("got %v, want %v", b.minCongestionWindow, minCongestionWindowPackets*newMaxDatagramSize) }
	if (initialCongestionWindowPackets*newMaxDatagramSize) != (b.congestionWindow) { t.Fatalf("got %v, want %v", b.congestionWindow, initialCongestionWindowPackets*newMaxDatagramSize) }
}

func TestSetMaxDatagramSizeClampsCongestionWindow(t *testing.T) {
	const oldMaxDatagramSize = congestion.ByteCount(1000)
	const newMaxDatagramSize = congestion.ByteCount(1400)

	b := NewBbrSender(DefaultClock{}, oldMaxDatagramSize, ProfileStandard)
	b.congestionWindow = b.minCongestionWindow + oldMaxDatagramSize
	b.recoveryWindow = b.minCongestionWindow + oldMaxDatagramSize

	b.SetMaxDatagramSize(newMaxDatagramSize)

	if (b.minCongestionWindow) != (b.congestionWindow) { t.Fatalf("got %v, want %v", b.congestionWindow, b.minCongestionWindow) }
	if (b.minCongestionWindow) != (b.recoveryWindow) { t.Fatalf("got %v, want %v", b.recoveryWindow, b.minCongestionWindow) }
}

func TestNewBbrSenderAppliesProfiles(t *testing.T) {
	testCases := []struct {
		name                                string
		profile                             Profile
		highGain                            float64
		highCwndGain                        float64
		congestionWindowGainConstant        float64
		numStartupRtts                      int64
		drainToTarget                       bool
		detectOvershooting                  bool
		bytesLostMultiplier                 uint8
		enableAckAggregationDuringStartup   bool
		expireAckAggregationInStartup       bool
		enableOverestimateAvoidance         bool
		reduceExtraAckedOnBandwidthIncrease bool
	}{
		{
			name:                                "standard",
			profile:                             ProfileStandard,
			highGain:                            defaultHighGain,
			highCwndGain:                        derivedHighCWNDGain,
			congestionWindowGainConstant:        2.0,
			numStartupRtts:                      roundTripsWithoutGrowthBeforeExitingStartup,
			bytesLostMultiplier:                 2,
			enableAckAggregationDuringStartup:   true,
			expireAckAggregationInStartup:       true,
			reduceExtraAckedOnBandwidthIncrease: true,
		},
		{
			name:                                "conservative",
			profile:                             ProfileConservative,
			highGain:                            2.25,
			highCwndGain:                        1.75,
			congestionWindowGainConstant:        1.75,
			numStartupRtts:                      2,
			drainToTarget:                       true,
			detectOvershooting:                  true,
			bytesLostMultiplier:                 1,
			enableOverestimateAvoidance:         true,
			reduceExtraAckedOnBandwidthIncrease: true,
		},
		{
			name:                              "aggressive",
			profile:                           ProfileAggressive,
			highGain:                          3.0,
			highCwndGain:                      2.25,
			congestionWindowGainConstant:      2.5,
			numStartupRtts:                    4,
			bytesLostMultiplier:               2,
			enableAckAggregationDuringStartup: true,
			expireAckAggregationInStartup:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := NewBbrSender(DefaultClock{}, congestion.InitialPacketSize, tc.profile)
			if (tc.profile) != (b.profile) { t.Fatalf("got %v, want %v", b.profile, tc.profile) }
			if (tc.highGain) != (b.highGain) { t.Fatalf("got %v, want %v", b.highGain, tc.highGain) }
			if (tc.highCwndGain) != (b.highCwndGain) { t.Fatalf("got %v, want %v", b.highCwndGain, tc.highCwndGain) }
			if (tc.congestionWindowGainConstant) != (b.congestionWindowGainConstant) { t.Fatalf("got %v, want %v", b.congestionWindowGainConstant, tc.congestionWindowGainConstant) }
			if (tc.numStartupRtts) != (b.numStartupRtts) { t.Fatalf("got %v, want %v", b.numStartupRtts, tc.numStartupRtts) }
			if (tc.drainToTarget) != (b.drainToTarget) { t.Fatalf("got %v, want %v", b.drainToTarget, tc.drainToTarget) }
			if (tc.detectOvershooting) != (b.detectOvershooting) { t.Fatalf("got %v, want %v", b.detectOvershooting, tc.detectOvershooting) }
			if (tc.bytesLostMultiplier) != (b.bytesLostMultiplierWhileDetectingOvershooting) { t.Fatalf("got %v, want %v", b.bytesLostMultiplierWhileDetectingOvershooting, tc.bytesLostMultiplier) }
			if (tc.enableAckAggregationDuringStartup) != (b.enableAckAggregationDuringStartup) { t.Fatalf("got %v, want %v", b.enableAckAggregationDuringStartup, tc.enableAckAggregationDuringStartup) }
			if (tc.expireAckAggregationInStartup) != (b.expireAckAggregationInStartup) { t.Fatalf("got %v, want %v", b.expireAckAggregationInStartup, tc.expireAckAggregationInStartup) }
			if tc.enableOverestimateAvoidance != b.sampler.IsOverestimateAvoidanceEnabled() { t.Fatalf("got %v, want %v", b.sampler.IsOverestimateAvoidanceEnabled(), tc.enableOverestimateAvoidance) }
			if (tc.reduceExtraAckedOnBandwidthIncrease) != (b.sampler.maxAckHeightTracker.reduceExtraAckedOnBandwidthIncrease) { t.Fatalf("got %v, want %v", b.sampler.maxAckHeightTracker.reduceExtraAckedOnBandwidthIncrease, tc.reduceExtraAckedOnBandwidthIncrease) }
			if (b.highGain) != (b.pacingGain) { t.Fatalf("got %v, want %v", b.pacingGain, b.highGain) }
			if (b.highCwndGain) != (b.congestionWindowGain) { t.Fatalf("got %v, want %v", b.congestionWindowGain, b.highCwndGain) }
		})
	}
}

func TestParseProfile(t *testing.T) {
	profile, err := ParseProfile("")
	if err != nil { t.Fatal(err) }
	if (ProfileStandard) != (profile) { t.Fatalf("got %v, want %v", profile, ProfileStandard) }

	profile, err = ParseProfile("Aggressive")
	if err != nil { t.Fatal(err) }
	if (ProfileAggressive) != (profile) { t.Fatalf("got %v, want %v", profile, ProfileAggressive) }

	_, err = ParseProfile("turbo")
	if err == nil || err.Error() != `unsupported BBR profile "turbo"` { t.Fatalf("got %v", err) }
}

func TestAckAggregationCwndClampedTo100ms(t *testing.T) {
	const mss = congestion.ByteCount(1400)
	b := NewBbrSender(DefaultClock{}, mss, ProfileStandard)
	b.isAtFullBandwidth = true
	bw := Bandwidth(100_000_000) // 100 Mbps = 12.5 MB/s -> 100ms = 1,250,000 bytes
	b.maxBandwidth.Update(bw, 1)
	b.sampler.maxAckHeightTracker.Reset(5_000_000, 1) // 5 MB raw height

	got := b.ackAggregationCwnd()
	want := bdpFromRttAndBandwidth(100*time.Millisecond, bw)
	if got != want {
		t.Fatalf("ackAggregationCwnd = %d, want clamped %d", got, want)
	}
}

func TestSingleRoundConservationRestoresPriorCwnd(t *testing.T) {
	const mss = congestion.ByteCount(1400)
	b := NewBbrSender(DefaultClock{}, mss, ProfileStandard)
	b.isAtFullBandwidth = true
	b.mode = bbrModeProbeBw
	b.minRtt = 25 * time.Millisecond
	b.maxBandwidth.Update(Bandwidth(400_000_000), 1)
	prior := congestion.ByteCount(2_500_000)
	b.congestionWindow = prior
	b.lastSentPacket = 100

	// Enter conservation on loss, with low instantaneous bytesInFlight.
	b.updateRecoveryState(80, true, false)
	b.bytesInFlight = 200 * mss
	b.calculateRecoveryWindow(10*mss, mss)
	if !b.InRecovery() {
		t.Fatal("expected sender to enter conservation on first loss")
	}
	if got := b.GetCongestionWindow(); got >= prior {
		t.Fatalf("expected conservation cwnd < prior (%d), got %d", prior, got)
	}

	// Next round-trip start clears conservation and restores priorCwnd even if another loss occurs.
	b.lastSentPacket = 200
	b.updateRecoveryState(105, true, true)
	if b.InRecovery() {
		t.Fatal("expected conservation to clear on round start")
	}
	if got := b.GetCongestionWindow(); got != prior {
		t.Fatalf("expected restored cwnd %d, got %d", prior, got)
	}
}
