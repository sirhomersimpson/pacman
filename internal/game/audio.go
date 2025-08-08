package game

import (
	"bytes"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

type SoundData struct {
	raw []byte
}

type AudioManager struct {
	ctx         *audio.Context
	pellet      *SoundData
	powerPellet *SoundData
	ghostEaten  *SoundData
	death       *SoundData
}

var (
	audioOnce sync.Once
	audioCtx  *audio.Context
)

func getAudioContext() *audio.Context {
	// Audio is DISABLED by default. Enable explicitly with PACMAN_ENABLE_AUDIO=1.
	if os.Getenv("PACMAN_DISABLE_AUDIO") == "1" {
		return nil
	}
	if os.Getenv("PACMAN_ENABLE_AUDIO") != "1" {
		return nil
	}
	audioOnce.Do(func() {
		audioCtx = audio.NewContext(44100)
	})
	return audioCtx
}

func NewAudioManager(soundsDir string) *AudioManager {
	if soundsDir == "" {
		soundsDir = "assets/sounds"
	}
	ctx := getAudioContext()
	am := &AudioManager{ctx: ctx}

	// Attempt to load each sound; if missing, synthesize a simple beep fallback
	if sd, _ := loadSoundData(soundsDir, "pellet.wav"); sd != nil {
		am.pellet = sd
	} else {
		am.pellet = &SoundData{raw: synthBeepWAV(44100, 60, 880)}
	}
	if sd, _ := loadSoundData(soundsDir, "power.wav"); sd != nil {
		am.powerPellet = sd
	} else {
		am.powerPellet = &SoundData{raw: synthBeepWAV(44100, 150, 660)}
	}
	if sd, _ := loadSoundData(soundsDir, "ghost.wav"); sd != nil {
		am.ghostEaten = sd
	} else {
		am.ghostEaten = &SoundData{raw: synthBeepWAV(44100, 200, 440)}
	}
	if sd, _ := loadSoundData(soundsDir, "death.wav"); sd != nil {
		am.death = sd
	} else {
		am.death = &SoundData{raw: synthBeepWAV(44100, 400, 220)}
	}
	return am
}

func loadSoundData(dir, file string) (*SoundData, error) {
	path := filepath.Join(dir, file)
	b, err := os.ReadFile(path)
	if err != nil {
		// Translate path errors to fs.PathError for clarity but return nil data
		if _, ok := err.(*fs.PathError); ok || os.IsNotExist(err) {
			return nil, err
		}
		return nil, err
	}
	return &SoundData{raw: b}, nil
}

func (am *AudioManager) play(sd *SoundData) {
	if am == nil || am.ctx == nil || sd == nil || len(sd.raw) == 0 {
		return
	}
	// Decode from bytes each time to allow overlapping plays
	stream, err := wav.Decode(am.ctx, bytes.NewReader(sd.raw))
	if err != nil {
		return
	}
	p, err := audio.NewPlayer(am.ctx, stream)
	if err != nil {
		return
	}
	_ = p.Rewind()
	p.Play()
}

func (am *AudioManager) PlayPellet()      { am.play(am.pellet) }
func (am *AudioManager) PlayPowerPellet() { am.play(am.powerPellet) }
func (am *AudioManager) PlayGhostEaten()  { am.play(am.ghostEaten) }
func (am *AudioManager) PlayDeath()       { am.play(am.death) }

// synthBeepWAV returns a minimal 16-bit PCM mono WAV of a sine beep.
func synthBeepWAV(sampleRate int, durationMs int, freq float64) []byte {
	numSamples := int(float64(sampleRate) * float64(durationMs) / 1000.0)
	// WAV header (44 bytes)
	byteRate := sampleRate * 2 // mono 16-bit
	blockAlign := 2
	dataSize := numSamples * 2
	totalSize := 44 + dataSize
	buf := make([]byte, totalSize)
	// RIFF header
	copy(buf[0:4], []byte{'R', 'I', 'F', 'F'})
	putLE32(buf[4:8], uint32(totalSize-8))
	copy(buf[8:12], []byte{'W', 'A', 'V', 'E'})
	// fmt chunk
	copy(buf[12:16], []byte{'f', 'm', 't', ' '})
	putLE32(buf[16:20], 16) // PCM chunk size
	putLE16(buf[20:22], 1)  // PCM format
	putLE16(buf[22:24], 1)  // channels
	putLE32(buf[24:28], uint32(sampleRate))
	putLE32(buf[28:32], uint32(byteRate))
	putLE16(buf[32:34], uint16(blockAlign))
	putLE16(buf[34:36], 16) // bits per sample
	// data chunk
	copy(buf[36:40], []byte{'d', 'a', 't', 'a'})
	putLE32(buf[40:44], uint32(dataSize))
	// samples
	amp := 0.25 // reduce volume
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		s := math.Sin(2 * math.Pi * freq * t)
		v := int16(s * 32767.0 * amp)
		off := 44 + i*2
		buf[off] = byte(v)
		buf[off+1] = byte(v >> 8)
	}
	return buf
}

func putLE16(b []byte, v uint16) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
}

func putLE32(b []byte, v uint32) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}
