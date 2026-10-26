#!/usr/bin/env python3
import asyncio
import os
import shutil
import subprocess
import sys
from pathlib import Path
import edge_tts

VOICE = "en-US-ChristopherNeural"
RATE = "+6%"
FONT = "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"

SCENES = [
    {
        "id": "01_overview",
        "start": 0.0,
        "end": 12.5,
        "hud_kicker": "GARUDA v1.4  |  SYSTEM COMMAND CENTER",
        "hud_title": "Deterministic Change-Control & Active Governance for AI Code",
        "vo": (
            "AI coding agents write software at a hundred miles an hour, "
            "but they forget your architecture the moment the prompt ends. "
            "Meet Garuda: the deterministic change-control and active enforcement plane for AI-written code."
        ),
    },
    {
        "id": "02_talon_contract",
        "start": 3.0,
        "end": 9.8,
        "hud_kicker": "GARUDA TALON  |  ACTIVE ARCHITECTURAL CONTRACT",
        "hud_title": "Mandatory Pre-Edit & Post-Edit Agent Execution Protocol (Epoch #52)",
        "vo": (
            "Instead of relying on an agent's memory, Garuda Talon binds Cursor, Claude Code, "
            "and Windsurf to a live architectural contract, requiring pre-edit blast-radius checks "
            "and post-edit verification on every single file."
        ),
    },
    {
        "id": "03_docs_graph",
        "start": 21.0,
        "end": 32.5,
        "hud_kicker": "SURFACE 04  |  DOCS, SPECS & GUARDRAILS GRAPH",
        "hud_title": "10 Documents -> 12 Extracted Normative Rules -> 11 Connected AST Symbols",
        "vo": (
            "Inside the Docs and Rules Graph, Garuda turns static markdown ADRs and YAML policies "
            "into live, compiler-verified guardrails, mapping every MUST, MUST NOT, and layer boundary "
            "directly to the exact functions and structs in your codebase."
        ),
    },
    {
        "id": "04_interactive_filter",
        "start": 79.0,
        "end": 87.5,
        "hud_kicker": "INTERACTIVE LINEAGE  |  3-COLUMN BIPARTITE GRAPH",
        "hud_title": "Click Any Policy or ADR to Isolate Exact AST Callers & Targets",
        "vo": (
            "Click any policy or guardrail card to dynamically isolate the exact rules "
            "and downstream code symbols it governs."
        ),
    },
    {
        "id": "05_ingest_modal",
        "start": 35.0,
        "end": 44.5,
        "hud_kicker": "LIVE INGESTION  |  ADR-0004 ZERO-TRUST AGENT GATE",
        "hud_title": "Extract Normative Rules & Link to Code Graph in Milliseconds",
        "vo": (
            "Drop in any architecture decision record, like ADR-0004, and Garuda extracts "
            "its normative rules and links them to your AST graph in milliseconds."
        ),
    },
    {
        "id": "06_talon_risk_hub",
        "start": 45.0,
        "end": 49.8,
        "hud_kicker": "SURFACE 05  |  GARUDA TALON & ARCHITECTURAL RISK HUB",
        "hud_title": "6/6 Rules Enforced  |  Cursor, Claude Code & Git Hooks Active  |  Auto-Repair",
        "vo": (
            "The Talon and Risk Hub gives engineering leaders one hundred percent enforcement visibility "
            "across Cursor, Claude Code, and Git hooks, while ranking your highest blast-radius hotspots "
            "and generating one-click self-repair prompts."
        ),
    },
    {
        "id": "07_change_review_proof",
        "start": 55.5,
        "end": 73.5,
        "hud_kicker": "SURFACE 06  |  CHANGE REVIEW WORKBENCH & RFC 6962 PROOF",
        "hud_title": "3-Plane Verification  |  Canonical SHA-256 Merkle Inclusion Receipt",
        "vo": (
            "In the Change Review Workbench, developers simulate the exact governance verdict, "
            "inspect the Talon pre-edit briefing, and verify the tamper-evident RFC 6962 Merkle inclusion proof "
            "with a single click."
        ),
    },
    {
        "id": "08_ledger_agents_closer",
        "start": 94.0,
        "end": 104.5,
        "hud_kicker": "SURFACES 07 & 08  |  EVIDENCE LEDGER, AGENT TIMELINE & ADR LINEAGE",
        "hud_title": "Don't Rely on Memory. Rely on Proof.  —  Garuda Software Intelligence",
        "vo": (
            "Every OpenTelemetry runtime trace and AI agent tool call is audited in real time, "
            "and every architectural decision is preserved forever. "
            "Don't rely on memory. Rely on proof, with Garuda."
        ),
    },
]

async def generate_tts(text: str, out_mp3: Path):
    communicate = edge_tts.Communicate(text, VOICE, rate=RATE)
    await communicate.save(str(out_mp3))

def get_duration(file_path: Path) -> float:
    res = subprocess.run(
        [
            "ffprobe", "-v", "error", "-show_entries", "format=duration",
            "-of", "default=noprint_wrappers=1:nokey=1", str(file_path)
        ],
        capture_output=True, text=True, check=True
    )
    return float(res.stdout.strip())

async def main():
    if len(sys.argv) < 2:
        print("Usage: python3 build_garuda_demo.py <path_to_screen_recording.mp4>")
        sys.exit(1)

    src_video = Path(sys.argv[1]).resolve()
    if not src_video.exists():
        print(f"Error: {src_video} not found.")
        sys.exit(1)

    work_dir = Path("garuda_video_build")
    if work_dir.exists():
        shutil.rmtree(work_dir)
    work_dir.mkdir(parents=True)

    segment_files = []

    for idx, scene in enumerate(SCENES, start=1):
        print(f"[{idx}/{len(SCENES)}] Processing {scene['id']}...")
        vo_path = work_dir / f"{scene['id']}_vo.mp3"
        await generate_tts(scene["vo"], vo_path)

        vo_dur = get_duration(vo_path) + 0.35
        raw_dur = scene["end"] - scene["start"]
        speed_factor = raw_dur / vo_dur

        seg_out = work_dir / f"{scene['id']}_seg.mp4"
        
        # Write text to files to prevent any shell/FFmpeg escaping issues with quotes
        kicker_file = work_dir / f"{scene['id']}_kicker.txt"
        kicker_file.write_text(scene["hud_kicker"], encoding="utf-8")
        title_file = work_dir / f"{scene['id']}_title.txt"
        title_file.write_text(scene["hud_title"], encoding="utf-8")

        vf_filter = (
            f"scale=1920:1080:force_original_aspect_ratio=decrease,"
            f"pad=1920:1080:(ow-iw)/2:(oh-ih)/2:color=#0d100d,"
            f"setpts={1.0/speed_factor:.4f}*PTS,"
            f"fps=30,"
            f"drawbox=x=40:y=965:w=1840:h=92:color=#0d100d@0.90:t=fill,"
            f"drawbox=x=40:y=965:w=6:h=92:color=#dcf56b@1.0:t=fill,"
            f"drawtext=fontfile={FONT}:textfile='{kicker_file.resolve()}':fontcolor=#dcf56b:fontsize=18:x=68:y=982,"
            f"drawtext=fontfile={FONT}:textfile='{title_file.resolve()}':fontcolor=#f4f6ef:fontsize=24:x=68:y=1014"
        )

        cmd = [
            "ffmpeg", "-y",
            "-ss", str(scene["start"]),
            "-to", str(scene["end"]),
            "-i", str(src_video),
            "-i", str(vo_path),
            "-filter_complex", f"[0:v]{vf_filter}[v];[1:a]apad=pad_dur=0.35[a]",
            "-map", "[v]", "-map", "[a]",
            "-t", f"{vo_dur:.2f}",
            "-c:v", "libx264", "-preset", "fast", "-crf", "18",
            "-c:a", "aac", "-b:a", "192k",
            str(seg_out)
        ]
        res = subprocess.run(cmd, capture_output=True, text=True)
        if res.returncode != 0:
            print(f"FFmpeg error in {scene['id']}:\n{res.stderr}")
            sys.exit(1)

        segment_files.append(seg_out)

    concat_list = work_dir / "concat.txt"
    concat_list.write_text("\n".join(f"file '{f.resolve()}'" for f in segment_files), encoding="utf-8")

    final_out = Path("garuda_investor_demo_voiceover.mp4")
    subprocess.run(
        [
            "ffmpeg", "-y", "-f", "concat", "-safe", "0",
            "-i", str(concat_list),
            "-c", "copy", str(final_out)
        ],
        check=True
    )
    print(f"\nDone! Created studio demo video: {final_out.resolve()}")

if __name__ == "__main__":
    asyncio.run(main())
