---

How KV Caching reduces inference cost by up to 10x
Prompt (KV) Caching: A Deep Dive Into 10x Cheaper AI Inference
KV Caching explained in detail
reduced latency of inference with caching (made by author in excalidraw)For non-members read here
We all know the basic unit of AI (LLMs) is tokens; you get billed on their usage, so the more you burn them, the higher your cost. The costs were reasonable when we only had a simple QnA chatbot interface with LLMs, but it all blew up in proportion when agentic engineering came into being. A long-running agentic session can burn millions in tokens & exhaust your limits or your bank balance (depending on whether you're on the API or a subscription).
So, how did the labs reduce compute costs for themselves and their customers without doing much? Well, the same old answer: Caching. Cache the tokens that are already computed so that the LLMs don't have to re-infer them on every new message or tool call.
Let's start with how LLMs work on a fundamental level.

---

Large Language Models as a White Box
LLMs are basically next-token predictors (quite brilliant ones tbh) or, as some say, "glorified autocomplete". The architecture of an LLM is composed of these components; everything you ever sent to an LLM goes into and comes out of this.
LLM architecture (made by author)LLMs are "stateless" in nature, so whenever you want to generate the next response, the entire previous conversation (responses, tool calls, and system prompt) is sent through the API call, and the model reprocesses that context from scratch to produce the next tokens.
So as the session grows, all of this data is processed again and again by the model, increasing the cost and latency; they also can't process everything that you sent to them indefinitely. Response quality degrades as the session data grows, and as a result LLMs cannot pay "attention" to what's important to the current query. That's why they have a context window and have to compact the conversation as it nears the context window limit.
It's also proven that a 1M context window performs worse than the traditional 250–400K window due to increased noise.
This reprocessing ties up GPUs and inherently increases the cost for the AI labs, which is then charged to you.
Let's see why this is the way it is, and then how caching fixes it.
A basic pseudo-code of an LLM (based on the above image) looks like this:
//taken from article by Sam Rose (see references)

prompt = "How does an LLM work?";

tokens = tokenizer(prompt);
while (true) {
    embeddings = embed(tokens);
    for ([attention, feedforward] of transformers) {
        embeddings = attention(embeddings);
        embeddings = feedforward(embeddings);
    }
    output_token = output(embeddings);
    if (output_token === END_TOKEN) {
        break;
    }
    tokens.push(output_token);
}

console.log(decode(tokens));
Let's talk briefly about all of these components:
Tokeniser
The tokenizer takes your prompt, breaks it into small chunks, and gives each unique chunk an integer ID called a "token." This is the fundamental unit LLMs operate on, as mentioned before. An LLM doesn't see the text that you send, but a stream of integers coming in and, based on math, outputs another stream of integers which get converted to matching text based on vector embeddings of the training data.
https://platform.openai.com/tokenizerEND_TOKEN defines when the inference stops. LLMs have a variety of "special" tokens they can output, one of which signals the end of a response, just like we have the EOF (end of file) when streaming our files. In the GPT-5 tokenizer, it's 199,999. There are many other such tokens, different for different models, telling them different states of their responses.
Embeddings
Embeddings are used to describe the tokens and their properties. Since tokens are just plain old integers, finding relationships based on their properties isn't naturally possible. 
That's where embeddings come in.
Embeddings are generated when a model is trained. These relationships between tokens (or original text) come out automatically when we try to train a model to produce a certain output based on a variety of inputs.
As mentioned earlier, each token has its own embedding describing its inherent properties for that training; it can be 1, 2, 3… n, any integer you can think of; these are called dimensions.
Each token starts with an array of length n with random embedding values, and as the training continues, these embeddings get pushed towards a suitable array state.
So each token can have n dimensions for its embeddings, and if we have m tokens, we get an m x n matrix of embeddings to work with during training and later inference (now you may guess why we have matrix multiplication in LLMs).
During training, each token gets moved within this space to be close to other, similar tokens. The more dimensions, the more complex and nuanced the LLM's representation of each token can be.
Transformer (In short)
"Attention is all you need" forms the basis of the transformer architecture and explains why LLMs have to reprocess everything to predict the next token.
Attention, as a concept, helps the model understand the relationship between the tokens in the prompt by moving their embeddings relative to each other in the n-dimensional space.
Internally, transformers generating text have two phases:
Prefill: the model reads the entire prompt in one pass and computes intermediate attention states for every token. This is compute-bound and scales with prompt length.
Decode: the model emits output tokens one at a time, each attending back over the whole prompt. (this is the main thing)

This is where the matrix multiplications happen to compute final tokens based on input tokens (everything happening with embeddings).
Matrix multiplication in a nutshell is:
M1 (3 x 2) * M2 (2 x 3) = MM (3 x 3)
Internally, after model training, embeddings are set and don't change during inference. These are WQ, WK, and WV weights.
WQ, WK, and WV are the sets of model weights and parameters (n-dimensional matrices) that become fixed and immutable during inference.
When you hear of a model having "1 trillion parameters," these are the numbers they're talking about (n).
For each token's embedding x, the model computes its Query, Key, and Value by projecting through these matrices.
maths formulae to calculate query, key and value matricesNow you may see why we call the caching in LLMs KV caching, or Key-Value caching.
So in a nutshell:
Self-attention works by having every token compare itself against every previous token, using the Query, Key, and Value vectors introduced above: a token's output is a weighted sum of the Values of all prior tokens, weighted by how well its Query matches each of their Keys.
A token's Key and Value depend only on that token and the tokens before it. So once the token's K and V are computed, they never change, no matter what gets appended later. That's what makes them cacheable.
There's a full step-by-step article explaining how these matrices are used to find the next token and generate responses by Sam Rose (see References [1]).
Now let's talk about caching.

---

KV Caching
Since you now know why we need caching and how it works, let's see some properties of KV caching.
Since a token's K/V depends on everything preceding it, changing a single token at position k changes the KV of every token from k onward, causing a cache-bust. So a cache hit is all-or-nothing up to the first difference; before should be pre-computed and fixed, after must be computed.
Put stable content (system prompt, tool definitions, etc.) first & volatile content at the end. Anything volatile placed early destroys the entire tail behind it.
The cache key will be a hash of the actual token IDs. Two prompts that mean the same thing but are worded differently share no prefix and get no hit. (we'll talk about this problem later)

Anthropic (Claude)
Claude uses explicit cache breakpoints. You mark a content block with cache_control, and everything up to that point becomes a cacheable prefix.
The pricing of cache read is evident here:
claude pricing with and without caching (by AI)Their TTL by default is 5 minutes but can be extended to 1 hour by passing an option {"ttl": "1h"} through an API request. Their cache hierarchy works in the order: tools -> system -> messages. Modifying any level breaks the cache and invalidates it.
OpenAI (GPT)
OpenAI caches automatically; no code changes required for the latest GPT-5 model family.
Their TTL is ~5–10 minutes for older models (up to 1 hour).
The GPT-5.6 family guarantees a 30-minute minimum retention. Their cache writes are free for pre-5.6 models but 1.25× on GPT-5.6 and later.
Overall, Anthropic provides more control over the caching, while OpenAI handles most of it for you.

---

Why Caching is needed for Agentic Workflows
A typical agentic loop can burn anywhere from 10–100k tokens, sometimes even more for longer-running reasoning tasks.
A typical agent step sends:
[ system prompt ]        <- stable, large (instructions, persona, rules)
[ tool definitions ]     <- stable, often large (many tools with schemas)
[ conversation so far ]  <- grows every step, but the prefix is stable
[ newest observation ]   <- the only genuinely new content
Without caching, you pay full price to reprocess a 50K-token context ten times. With caching, you pay the write cost once and then 10% for every subsequent step.
Latency also drops as models don't have to reprocess old tokens to provide new outputs.
Here's the math on an example. Take a 100,000-token context, queried repeatedly, with Claude Opus 4.8 at $5/MTok:
cost comparison between caching and no-caching agentic sessionThat's a ~90% reduction on every step after the first. Over a 20-step agent run, the cached version costs roughly $0.625 + 19×$0.05 = $1.58 versus 20×$0.50 = $10.00 uncached.

---

Designing a Custom Cache system for Long-Running Agent Sessions
Basic provider caching doesn't work when you have a lot of changing needs and sessions.
The first thought could be to have a simple setup: hash the prompt & store it with the result and serve from it when the same prompt comes in. For simple QnA sessions it may work, but not for agentic sessions.
A few design decisions to consider when designing for agentic workflows:
Cost is climbing because every step re-sends and reprocesses an ever-longer context. A response cache stores the cheap thing (the final answer) & ignores the expensive thing (reprocessing the prefix).
Each step appends a fresh tool result or observation, so the full prompt differs every time. A key over the whole prompt essentially never hits.
To rescue the hit rate, you will reach for semantic matching: "same question, different wording, same answer." But handing an agent a stale near-match can send it down a wrong branch and corrupt the rest of the session. This is a crucial tradeoff that you need to understand and design around, because it's not a "one size fits all" type of scenario.

What we're actually building (and what we're not)
We know every AI lab provider already stores the KV state, hashes incoming prompts to find matches, and serves hits at the discounted rate. 
That's a black box we don't care about.
What we're designing is the agent-side layer that sits on top of the API and makes the provider's cache actually hit. It's a two-layer system: the provider supplies the mechanism, and this design supplies the policy that decides whether the mechanism ever pays off.
Follow the cost curve.
A typical agentic loop follows this cycle:
Think -> Call tool -> Observe -> Think …
And every iteration, the model re-processes the whole previous step's data. So, what's redundant is everything up to the previous turn, because it's the same byte-for-byte.
Design
Instead of caching answers keyed by prompt meaning (semantics), use the provider's prefix cache: the model's intermediate attention state (the KV cache) for the shared prefix, keyed by an exact token hash:
Key: a hash of the token prefix, computed incrementally per segment. The exact token sequence, not its meaning; it's deterministic & not error-prone.
Value: the KV vectors already computed for that prefix, stored inside the provider's inference layer.
On every step, the whole prefix is a hit; you only pay to prefill the new suffix.

So you only pay for cache writes (first time the tokens enter the provider's layer), and then the subsequent discounted rates for cache reads. Your design makes sure the cache keeps hitting on the provider side so that you get the discounted rates as the session keeps on growing.
If history grows forever, the bill still has a quadratic shape with a much smaller constant; once you add compaction or sliding windows, the practical curve gets much closer to linear.
This is how the cost of uncached vs cached with context-managed sessions looks.
The uncached total climbs quadratically because each step reprocesses a longer prefix. Prefix caching discounts that repeated work; compaction is what keeps the prefix from growing without bound.High Level Design
HLD of caching system (made by author)Read the diagram as one request flowing left to right.
The Agent Runtime is doing its normal loop: think, call a tool, observe, ask the model again, but before each model call it hands the conversation to the Context Assembler. That assembler has one job: keep the request in a cache-friendly shape. Stable content goes into the prefix & the newest observation stays in the volatile suffix.
From there, the Inference Engine receives the full prompt. The suffix is always fresh work, but the prefix can be matched against the Provider Prompt Cache by exact token hash. If a cache is hit, the provider reuses the cached KV state for the prefix and only pre-fills the new suffix; otherwise the provider computes the prefix & writes the resulting KV state back into its own cache. The agent only earns hits by making the prefix byte-identical across calls.
The Context Manager runs when the conversation gets too large: it compacts, truncates, or offloads old context, knowing that this starts a new cache epoch. The Guarded Response Cache is even narrower: it is only for safe, read-only sub-calls where replaying an exact previous answer is needed & fine. It's never the main mechanism for the agentic loop. (can be dropped as it is optional)
The load-bearing idea is the prefix/suffix split. The Context Assembler is built to keep the stable prefix immutable and push everything volatile into the suffix. The prompt cache is provider-owned; the agent owns the ordering, breakpoints, TTL choice, and telemetry that make it hit, whereas the response cache is separate & optional, so it sits outside the main design & is only needed on a few niche occasions.
The tandem between the assembler and the compactor is the real deal. Caching and compaction pull against each other. Compaction breaks the cache to shrink the prefix. Let the cache run between compactions & compact only when the prefix has grown enough to justify the reset.

---

Guarded Response Cache: its earned use-cases
As the separate & optional outer layer in the diagram, a response cache genuinely helps for idempotent, read-only sub-calls: "summarize this document," "classify this text", where the model, prompt, decoding settings, and input are pinned so tightly that replaying a prior answer is fine.
Prefer exact-input keys by default and reserve semantic matching for narrow, audited cases where a stale near-match won't corrupt the agent's next step.
Where prefix caching cuts the recurring cost of the main loop, the response cache instead picks off repeated sub-work at the edges.
Overall, it's fascinating to see how AI inference lets us maintain low cost and latency for agentic use-cases.
Hope this helps ^-^

---

References
https://ngrok.com/blog/prompt-caching [Prompt caching article by Sam Rose.]
Anthropic - Prompt caching documentation [cache-write vs. cache-read pricing multipliers, minimum token thresholds, TTL options, cache-prefix hierarchy, and invalidation rules.]
OpenAI - Prompt caching guide [automatic prefix caching, minimum token threshold, cached-input discount, and retention behavior.]

---

Support writers. All our nonprofit's offerings here.
Click here to subscribe to the IT Chronicles newsletter.
Click for Wordsmith, Mystery Writing, Write Like Stephen King, more
By the EIC Susan Brearley with Ideogram

---