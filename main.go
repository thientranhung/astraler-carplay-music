package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	version   = "1.0.0"
	ttsDir    = "assets/tts"
	bgmDir    = "assets/bg-music"
	outputDir = "_output"
)

var audioExts = []string{".mp3", ".wav", ".m4a"}

func findAudio(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("không đọc được thư mục '%s': %w", dir, err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		for _, ae := range audioExts {
			if ext == ae {
				files = append(files, filepath.Join(dir, e.Name()))
				break
			}
		}
	}
	return files, nil
}

func mix(ttsPath, bgPath, outPath string, bgmVolume, ttsVolume float64, delayMs int) error {
	filter := fmt.Sprintf(
		"[0]volume=%.1fdB[bg];[1]volume=%.1fdB,adelay=%d|%d[tts];[bg][tts]amix=inputs=2:duration=first:dropout_transition=0[out]",
		bgmVolume, ttsVolume, delayMs, delayMs,
	)
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", bgPath,
		"-i", ttsPath,
		"-filter_complex", filter,
		"-map", "[out]",
		"-b:a", "320k",
		outPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg lỗi: %s", string(out))
	}
	return nil
}

func run(bgmVolume, ttsVolume float64, delayMs int) error {
	for _, dir := range []string{ttsDir, bgmDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("không tìm thấy thư mục '%s'. Hãy chạy từ thư mục gốc của project", dir)
		}
	}

	ttsFiles, err := findAudio(ttsDir)
	if err != nil {
		return err
	}
	bgFiles, err := findAudio(bgmDir)
	if err != nil {
		return err
	}

	if len(ttsFiles) == 0 {
		return fmt.Errorf("không tìm thấy file audio trong '%s/'. Hỗ trợ: .mp3, .wav, .m4a", ttsDir)
	}
	if len(bgFiles) == 0 {
		return fmt.Errorf("không tìm thấy file audio trong '%s/'. Hỗ trợ: .mp3, .wav, .m4a", bgmDir)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("không tạo được thư mục output: %w", err)
	}

	total := len(ttsFiles) * len(bgFiles)
	fmt.Fprintf(os.Stderr, "%d TTS × %d BG = %d file  [bgm %+.0fdB | tts %+.0fdB | delay %.1fs]\n\n",
		len(ttsFiles), len(bgFiles), total, bgmVolume, ttsVolume, float64(delayMs)/1000)

	for _, bg := range bgFiles {
		bgName := strings.TrimSuffix(filepath.Base(bg), filepath.Ext(bg))
		for _, tts := range ttsFiles {
			ttsName := strings.TrimSuffix(filepath.Base(tts), filepath.Ext(tts))
			out := filepath.Join(outputDir, ttsName+"+"+bgName+".mp3")
			if err := mix(tts, bg, out, bgmVolume, ttsVolume, delayMs); err != nil {
				return fmt.Errorf("%s: %w", out, err)
			}
			fmt.Printf("  ✓ %s\n", filepath.Base(out))
		}
	}

	fmt.Fprintf(os.Stderr, "\nXong. %d file trong '%s/'\n", total, outputDir)
	return nil
}

func main() {
	bgmVolume := flag.Float64("bgm-volume", -3, "Điều chỉnh âm lượng nhạc nền, đơn vị dB (mặc định: -3)")
	ttsVolume := flag.Float64("tts-volume", 0, "Điều chỉnh âm lượng giọng TTS, đơn vị dB (mặc định: 0)")
	delay := flag.Float64("delay", 0.5, "Số giây nhạc nền chạy trước khi TTS bắt đầu (mặc định: 0.5)")
	showVersion := flag.Bool("version", false, "Hiển thị version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "astra-carplay-music %s\n\n", version)
		fmt.Fprintf(os.Stderr, "Ghép tất cả file TTS × BG music, xuất ra %s/.\n", outputDir)
		fmt.Fprintf(os.Stderr, "Tự động quét %s/ và %s/, tạo N×M file output.\n\n", ttsDir, bgmDir)
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  astra-carplay-music\n")
		fmt.Fprintf(os.Stderr, "  astra-carplay-music --bgm-volume -6 --tts-volume 3\n")
		fmt.Fprintf(os.Stderr, "  astra-carplay-music --delay 1.0 --bgm-volume -8\n\n")
		fmt.Fprintf(os.Stderr, "Thư mục:\n")
		fmt.Fprintf(os.Stderr, "  TTS input  : %s/\n", ttsDir)
		fmt.Fprintf(os.Stderr, "  BG input   : %s/\n", bgmDir)
		fmt.Fprintf(os.Stderr, "  Output     : %s/\n\n", outputDir)
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("astra-carplay-music %s\n", version)
		os.Exit(0)
	}

	if err := run(*bgmVolume, *ttsVolume, int(*delay*1000)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
