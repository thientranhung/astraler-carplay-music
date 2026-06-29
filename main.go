package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var version = "dev"

const (
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

func mix(ttsPath, bgPath, outPath string, bgmVolume, ttsVolume float64, delayMs int, duration, fadeOut float64) error {
	mixChain := "amix=inputs=2:duration=first:dropout_transition=0"
	if fadeOut > 0 && duration > 0 {
		fadeStart := duration - fadeOut
		mixChain += fmt.Sprintf(",afade=t=out:st=%.3f:d=%.3f", fadeStart, fadeOut)
	}
	filter := fmt.Sprintf(
		"[0]volume=%.1fdB[bg];[1]volume=%.1fdB,adelay=%d|%d[tts];[bg][tts]%s[out]",
		bgmVolume, ttsVolume, delayMs, delayMs, mixChain,
	)
	args := []string{"-y", "-i", bgPath, "-i", ttsPath, "-filter_complex", filter, "-map", "[out]", "-b:a", "320k"}
	if duration > 0 {
		args = append(args, "-t", fmt.Sprintf("%.3f", duration))
	}
	args = append(args, outPath)

	if out, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg lỗi: %s", string(out))
	}
	return nil
}

func doctor() {
	type check struct {
		name    string
		command string
		args    []string
		install string
	}
	checks := []check{
		{"ffmpeg", "ffmpeg", []string{"-version"}, "brew install ffmpeg"},
	}

	ok := true
	for _, c := range checks {
		out, err := exec.Command(c.command, c.args...).Output()
		if err != nil {
			fmt.Printf("  ✗ %-10s não tìm thấy  →  %s\n", c.name, c.install)
			ok = false
			continue
		}
		// Lấy dòng đầu tiên của output làm version string
		firstLine := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
		fmt.Printf("  ✓ %-10s %s\n", c.name, firstLine)
	}

	fmt.Println()
	if ok {
		fmt.Println("Tất cả dependencies OK. Tool sẵn sàng sử dụng.")
	} else {
		fmt.Fprintln(os.Stderr, "Còn thiếu dependencies. Cài đặt theo hướng dẫn ở trên rồi chạy lại.")
		os.Exit(1)
	}
}

func run(bgmVolume, ttsVolume float64, delayMs int, duration, fadeOut float64) error {
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
	durStr := "full"
	if duration > 0 {
		durStr = fmt.Sprintf("%.1fs", duration)
	}
	fadeStr := "none"
	if fadeOut > 0 {
		fadeStr = fmt.Sprintf("%.1fs", fadeOut)
	}
	fmt.Fprintf(os.Stderr, "%d TTS × %d BG = %d file  [bgm %+.0fdB | tts %+.0fdB | delay %.1fs | duration %s | fadeout %s]\n\n",
		len(ttsFiles), len(bgFiles), total, bgmVolume, ttsVolume, float64(delayMs)/1000, durStr, fadeStr)

	for _, bg := range bgFiles {
		bgName := strings.TrimSuffix(filepath.Base(bg), filepath.Ext(bg))
		for _, tts := range ttsFiles {
			ttsName := strings.TrimSuffix(filepath.Base(tts), filepath.Ext(tts))
			out := filepath.Join(outputDir, ttsName+"_"+bgName+".mp3")
			if err := mix(tts, bg, out, bgmVolume, ttsVolume, delayMs, duration, fadeOut); err != nil {
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
	duration := flag.Float64("duration", 0, "Độ dài file output tính bằng giây (mặc định: full)")
	fadeOut  := flag.Float64("fade-out", 0, "Fade out cuối bài, tính bằng giây (mặc định: không có)")
	showVersion := flag.Bool("version", false, "Hiển thị version")
	showDoctor  := flag.Bool("doctor", false, "Kiểm tra dependencies cần thiết")

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
		fmt.Fprintf(os.Stderr, "\nChạy --doctor để kiểm tra dependencies.\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("astra-carplay-music %s\n", version)
		os.Exit(0)
	}

	if *showDoctor {
		fmt.Printf("astra-carplay-music %s — doctor\n\n", version)
		doctor()
		os.Exit(0)
	}

	if err := run(*bgmVolume, *ttsVolume, int(*delay*1000), *duration, *fadeOut); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
