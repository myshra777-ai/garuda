# Token Telemetry from a Single Paired Codex CLI Run: A Pilot Study

Status: pilot. One valid paired run. Not a benchmark result.

## 1. Abstract

We ran one paired trial of Codex CLI answering a 3-hop question about the `SQLModel` class in a Python repository (sqlmodel), once without Garuda and once with Garuda's blast-radius tool available. Nine of ten planned trials per arm produced no data, most likely because of API rate limits. The single valid pair does not support a conclusion. The raw arm used one tool call and reported 170 function entities, which the Garuda output does not support. The Garuda arm used two tool calls and reported 27 function entities, with two call chains. Total tokens were 1,399,216 for the raw arm and 109,282 for the Garuda arm. The raw arm's cached-input count (1,304,320) and the Garuda arm's (83,968) reflect different cache states, so the difference cannot be attributed to the tool. We report this as a pilot to motivate a controlled study with a verified no-tool control, repeated trials, and blind scoring.

## 2. Methodology

**Task.** Answer: how many function entities are reachable from `SQLModel` within three hops, with call chains.

**Environment.** Codex CLI `exec`, ephemeral sessions, local checkout of the sqlmodel repository. Model identity was not recorded in the telemetry, so it is unverified across arms.

**Arms.**
- *Raw:* the question only. Repository rules were not confirmed to be disabled in the run that produced the telemetry. The raw run's tool usage (1 call) shows tools were available.
- *Garuda:* the question plus an instruction to call `garuda.blast_radius` at depth 3 before answering. This arm used 2 tool calls.

**Measurement.** Token counts come from the final usage record in each run's JSON event stream: uncached input, cached input, output, and reasoning output. Total is the sum reported by the harness. Tool and search calls were counted per run.

**Trials.** A script ran 10 trials per arm. Trials 2–10 in the raw arm and trials 2–10 in the Garuda arm returned zero telemetry, consistent with API rate-limit failures. Only trial 1 of each arm produced data. Failed trials are reported as failures, not omitted from the count of attempts.

**Scoring.** Answers were compared with each other and with the Garuda output. Scoring was not blind, and the person who built the tool also judged the answers. This is a known limitation.

## 3. Token Telemetry Results

Single paired run (trial 1 per arm). Nine other trials per arm produced no data.

| Arm | Tool calls | Uncached input | Cached input | Output | Reasoning | Total | Answer |
|---|---|---|---|---|---|---|---|
| Raw (control) | 1 | 86,754 | 1,304,320 | 5,802 | 2,340 | 1,399,216 | 170 function entities, 2 chains |
| Garuda | 2 | 24,747 | 83,968 | 443 | 124 | 109,282 | 27 function entities, 2 chains |

**Observed difference.** Total tokens: 1,399,216 vs. 109,282, a difference of 1,289,934 tokens, or 92.2% lower in the Garuda arm (the earlier figure of 92.1% used a slightly different rounding). Cached input: 1,304,320 vs. 83,968.

**Why this is not yet a reduction claim.**
- *Different tool use.* The arms did not differ only by the Garuda instruction. The raw arm also made a tool call.
- *Different cache state.* The cached counts depend on what was already in the cache when each run started. One run may have benefited from prior warming. Cache state at start was not recorded.
- *Different answers.* The two arms answered with different counts (170 vs. 27). Neither was verified against a reference, so it is unknown which, if either, is correct.
- *One pair.* With n = 1 per arm, there is no variance estimate and no confidence interval.

## 4. Analysis of the Context Tax

The hypothesis behind this pilot is that an agent without a structural index spends tokens re-reading large amounts of context to answer graph-shaped questions, and that this cost shows up mainly as cached input re-read across turns. The pilot is consistent with that hypothesis: the raw run's cached input is about 15 times the Garuda run's. It is also consistent with cache warmth alone, which would produce the same pattern without any difference in method.

The pilot cannot separate the two explanations. To do that, the study needs:

1. Cache state recorded at the start of each run, or runs spaced so the cache is cold for each arm.
2. Uncached input and output reported separately from cached input, since the billing rates differ.
3. Tool and search call counts per run, so the number of re-reads is measured directly.

Until those are in place, "context tax" is a label for an observed cost, not a demonstrated mechanism.

## 5. Conclusion

This pilot does not show that graph grounding reduces tokens or hallucinations, and it does not show that graph grounding is required for agentic workflows.

What it does show:
- A paired run on this question produced very different token totals and different answers.
- The raw arm's answer (170 entities) is not consistent with the structured tool output (27 functions), so at least one arm answered incorrectly or answered a different question.
- The measurement pipeline works end to end and captures per-run usage.

**Next steps for a valid study:**
1. Make the control arm tool-free, and verify that from the logs.
2. Run at least 20 valid trials per arm, with retries and backoff for rate limits, and report failures.
3. Pin and record the model and the cache state at the start of each run.
4. Score answers blind against a reference that a human verifies from the source code.
5. Report uncached input, cached input, output, and reasoning tokens separately, with confidence intervals.

## Appendix: Raw telemetry

```
Arm,Trial,Input_Uncached,Input_Cached,Output_Tokens,Reasoning_Tokens,Total_Tokens,Tool_Calls,Search_Calls
raw,1,86754,1304320,5802,2340,1399216,1,0
raw,2-10,0,0,0,0,0,0-1,0
garuda,1,24747,83968,443,124,109282,2,0
garuda,2-10,0,0,0,0,0,0,0
```

Trials 2–10 in both arms returned zero telemetry and are treated as failed attempts.
