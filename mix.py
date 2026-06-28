#!/usr/bin/env python3
"""
astra-mix-sound — Ghép tất cả file TTS × BG music, xuất ra _output/.

Tự động quét assets/tts/ và assets/bg-music/, tạo N×M file output.

Examples:
  python mix.py
  python mix.py --bgm-volume -6 --tts-volume 3
  python mix.py --delay 1.0 --bgm-volume -8
"""

import argparse
import os
import sys
import glob
from pydub import AudioSegment

__version__ = "1.0.0"

TTS_DIR = "assets/tts"
BGM_DIR = "assets/bg-music"
OUTPUT_DIR = "_output"

# ANSI color — chỉ bật khi output là terminal thật
_IS_TTY = sys.stdout.isatty()
GREEN  = "\033[32m" if _IS_TTY else ""
RED    = "\033[31m" if _IS_TTY else ""
DIM    = "\033[2m"  if _IS_TTY else ""
RESET  = "\033[0m"  if _IS_TTY else ""


def err(msg: str) -> None:
    """In lỗi ra stderr."""
    print(f"{RED}Error:{RESET} {msg}", file=sys.stderr)


def find_audio(directory: str) -> list[str]:
    exts = ("*.mp3", "*.wav", "*.m4a")
    files = []
    for ext in exts:
        files.extend(glob.glob(os.path.join(directory, ext)))
        files.extend(glob.glob(os.path.join(directory, ext.upper())))
    return sorted(set(files))


def mix(tts_path: str, bg_path: str, output_path: str,
        bgm_volume: float, tts_volume: float, delay_ms: int) -> None:
    tts = AudioSegment.from_file(tts_path).set_channels(2) + tts_volume
    bg  = AudioSegment.from_file(bg_path).set_channels(2) + bgm_volume
    final = bg.overlay(tts, position=delay_ms)
    final.export(output_path, format="mp3", bitrate="320k")
    duration = len(final) / 1000
    print(f"  {GREEN}✓{RESET} {os.path.basename(output_path)} {DIM}({duration:.1f}s){RESET}")


def main() -> None:
    parser = argparse.ArgumentParser(
        prog="astra-mix-sound",
        description=__doc__,
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=(
            "Thư mục:\n"
            f"  TTS input  : {TTS_DIR}/\n"
            f"  BG input   : {BGM_DIR}/\n"
            f"  Output     : {OUTPUT_DIR}/\n"
        ),
    )
    parser.add_argument(
        "--bgm-volume", type=float, default=-3, metavar="dB",
        help="Điều chỉnh âm lượng nhạc nền, đơn vị dB (mặc định: -3)",
    )
    parser.add_argument(
        "--tts-volume", type=float, default=0, metavar="dB",
        help="Điều chỉnh âm lượng giọng TTS, đơn vị dB (mặc định: 0)",
    )
    parser.add_argument(
        "--delay", type=float, default=0.5, metavar="giây",
        help="Số giây nhạc nền chạy trước khi TTS bắt đầu (mặc định: 0.5)",
    )
    parser.add_argument(
        "--version", action="version", version=f"%(prog)s {__version__}",
    )
    args = parser.parse_args()

    # Validate sớm trước khi xử lý
    missing = [d for d in (TTS_DIR, BGM_DIR) if not os.path.isdir(d)]
    if missing:
        for d in missing:
            err(f"Không tìm thấy thư mục '{d}'. Hãy chạy từ thư mục gốc của project.")
        sys.exit(1)

    tts_files = find_audio(TTS_DIR)
    bg_files  = find_audio(BGM_DIR)

    if not tts_files:
        err(f"Không tìm thấy file audio trong '{TTS_DIR}/'. Hỗ trợ: .mp3, .wav, .m4a")
        sys.exit(1)
    if not bg_files:
        err(f"Không tìm thấy file audio trong '{BGM_DIR}/'. Hỗ trợ: .mp3, .wav, .m4a")
        sys.exit(1)

    os.makedirs(OUTPUT_DIR, exist_ok=True)
    delay_ms = int(args.delay * 1000)
    total = len(tts_files) * len(bg_files)

    # Progress/info → stderr để không làm bẩn stdout
    print(
        f"{len(tts_files)} TTS × {len(bg_files)} BG = {total} file  "
        f"{DIM}[bgm {args.bgm_volume:+g}dB | tts {args.tts_volume:+g}dB | delay {args.delay}s]{RESET}",
        file=sys.stderr,
    )
    print(file=sys.stderr)

    for bg_path in bg_files:
        bg_name = os.path.splitext(os.path.basename(bg_path))[0]
        for tts_path in tts_files:
            tts_name = os.path.splitext(os.path.basename(tts_path))[0]
            out = os.path.join(OUTPUT_DIR, f"{tts_name}+{bg_name}.mp3")
            mix(tts_path, bg_path, out, args.bgm_volume, args.tts_volume, delay_ms)

    print(file=sys.stderr)
    print(f"Xong. {total} file trong '{OUTPUT_DIR}/'", file=sys.stderr)


if __name__ == "__main__":
    main()
