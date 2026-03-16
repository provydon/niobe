# How I Built Niobe: An AI Waitress with Gemini Live and Google Cloud

**Disclaimer:** *I created this blog post for the purposes of entering the Gemini Live Agent Challenge hackathon. When sharing on social media, use the hashtag #GeminiLiveAgentChallenge.*

---

## What is Niobe?

**Niobe** is an AI waitress for restaurants. Restaurant owners upload their menu (even as a photo), get a shareable link, and customers can talk to the waitress by voice—ask about the menu, place orders, and have a natural conversation. No app install required; everything runs in the browser.

I built it using **Google’s Gemini** (including the **Gemini Live API** for real-time voice) and **Google Cloud** for deployment. Here’s how I put it together.

---

## Two places where Gemini powers the product

I use Gemini in two different ways:

### 1. Gemini API: turning menu images into structured data

Restaurant owners don’t have to type their menu. They upload images (photos of the menu, PDFs, etc.). The **Laravel** backend sends those images to the **Gemini API** (via the Laravel AI package with the Gemini driver). Gemini returns structured text and JSON—dishes, categories, prices—and I store that in **PostgreSQL**. So the “brain” of the waitress (menu + context) is partly built by Gemini from images. This is configured in Laravel with `GEMINI_API_KEY` and the default provider in `config/ai.php`:

```php
// config/ai.php
'default' => 'gemini',
'default_for_images' => 'gemini',

'providers' => [
    'gemini' => [
        'driver' => 'gemini',
        'key' => env('GEMINI_API_KEY'),
    ],
    // ...
],
```

Menu extraction then calls Gemini with a prompt and the uploaded image attachments; the response is parsed as JSON and saved to the database.

### 2. Gemini Live API: the voice conversation

The real-time voice experience is powered by the **Gemini Live API**. My **Go voice agent** doesn’t do speech-to-text or text-to-speech itself; it acts as a **proxy** between the browser and Gemini Live:

- The customer opens the “Talk” page and connects via **WebSocket** to the Go service (`/live?niobe=<slug>`).
- The Go agent loads the waitress and menu from the **same PostgreSQL database** Laravel uses, builds a system instruction and tool definitions, and opens a **Gemini Live** session using the Google GenAI SDK.
- Audio flows both ways: browser ↔ Go agent ↔ Gemini Live. The model speaks and listens in real time, with natural turn-taking and interrupt handling.

When the model decides to take an action (e.g. “place order”), it sends a **tool call** to the agent. The Go service runs **LocalNiobeTools**: it writes to the database (e.g. `waitress_action_logs`), can send email or fire webhooks, and returns the result back to Gemini. The model then confirms to the user in voice. So: **Gemini Live** = voice + reasoning; **Go + PostgreSQL** = tools and persistence.

The Go agent connects to Gemini Live with the Google GenAI SDK and wires up system instruction, tools, and audio config:

```go
// agent/live/google.go (simplified)
client, _ := genai.NewClient(ctx, &genai.ClientConfig{
    APIKey:      cfg.GetAPIKey(),
    HTTPOptions: httpOpts,
})
model := "gemini-2.5-flash-native-audio-preview-12-2025"

connectConfig := &genai.LiveConnectConfig{
    ResponseModalities: []genai.Modality{genai.ModalityAudio},
    SpeechConfig: &genai.SpeechConfig{
        VoiceConfig: &genai.VoiceConfig{
            PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{VoiceName: "Aoede"},
        },
    },
    SystemInstruction: &genai.Content{
        Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
        Role:  genai.RoleUser,
    },
    Tools: tools,
}
sess, _ := client.Live.Connect(ctx, model, connectConfig)
```

From there, a proxy bridges the browser WebSocket and this session so audio and tool calls flow both ways.

---

## Why Google Cloud?

I run the app on **Google Cloud** so that the Laravel app and the Go agent can share one **Cloud SQL (PostgreSQL)** instance. The agent is deployed as a container (e.g. **Cloud Run**), and the web app is deployed via **Terraform** and **Cloud Build** in the `deploy/` folder. Same VPC and database mean low latency and a single source of truth for menus, waitresses, and action logs. Configuration is via environment variables—no secrets in code:

```bash
# Agent (Go) – example .env
GEMINI_API_KEY=your_key
DATABASE_URL=postgres://user:password@/niobe?host=/cloudsql/PROJECT:REGION:INSTANCE
PORT=8080
APP_URL=https://your-laravel-app.run.app

# Optional: use Vertex AI instead of Gemini API
# GOOGLE_GENAI_USE_VERTEXAI=true
```

---

## Architecture in one picture

- **Browser** → **Laravel** (HTTPS) for dashboard and menu upload; Laravel uses **Gemini API** for menu extraction and writes to **PostgreSQL**.
- **Browser** → **Go agent** (WebSocket) for voice; the agent reads from **PostgreSQL**, talks to **Gemini Live API**, and runs tools (DB, email, webhooks) in-process.

So: one database, two Gemini touchpoints (Gemini API for menus, Gemini Live for voice), and Google Cloud to host and connect it all.

---

## What I’d do next

I’d add more tool types (e.g. table reservation, kitchen display), support for more languages, and tighter integration with POS systems. The current stack—Laravel + Vue/Inertia for the app, Go for the voice proxy, Gemini for vision and live voice, and Google Cloud for deployment—gives me a clear path to scale.

If you want to see the code or run it yourself, check out the repo and the [architecture doc](https://github.com/your-org/niobe-project/blob/main/docs/ARCHITECTURE.md) for diagrams and data flows.

---

*This post was created for the purposes of entering the Gemini Live Agent Challenge hackathon. #GeminiLiveAgentChallenge*
