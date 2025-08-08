package game

import "testing"

// This test ensures the audio manager initializes and can be called even if
// no sound files exist, without panicking.
func TestAudioManagerNoAssets(t *testing.T) {
	t.Setenv("PACMAN_DISABLE_AUDIO", "1")
	am := NewAudioManager("/nonexistent/path")
	am.PlayPellet()
	am.PlayPowerPellet()
	am.PlayGhostEaten()
	am.PlayDeath()
}
