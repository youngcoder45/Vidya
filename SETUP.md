# 🏫 Vidya — Setup Guide for Absolute Beginners

This guide takes you from a **completely empty computer** to seeing the Vidya app
**running in your browser** with real data (students, fees, dashboard numbers).

You will install three things:

| # | What | Why |
|---|------|-----|
| 1 | **Docker** | Runs the backend (the brain) + its database all at once |
| 2 | **Flutter** | The app itself (login screen, dashboard) |
| 3 | **A browser** (Chrome/Firefox/Edge) | Where you'll see the app — you almost certainly already have one |

> ⏱️ **Total time:** about 30–60 minutes, mostly waiting for downloads.
> 💾 **Disk space needed:** about 6–8 GB free.
> 🧠 **Skill needed:** none. Just follow the steps in order and copy-paste the commands.

---

## 📦 Step 0 — Unzip the code

1. Put the zip file somewhere easy, like your **Desktop** or `Documents`.
2. **Extract** it (right-click → *Extract All* on Windows, double-click on Mac, or `unzip` on Linux).
3. The extracted folder will contain: `backend`, `mobile`, `deploy`, `docs`, `SETUP.md`, and other files.
4. Open a **terminal** (see below) and go into that folder.

**How to open a terminal:**
- **Windows:** Press `Win` key → type `PowerShell` → press Enter
- **Mac:** Press `Cmd + Space` → type `Terminal` → press Enter
- **Linux:** `Ctrl + Alt + T`

Then type this (replace the path with where you extracted it — drag-and-drop the folder onto the terminal window to type the path automatically):

```bash
cd "C:\path\to\extracted\folder"      # Windows example
cd ~/Desktop/extracted-folder         # Mac / Linux example
```

> 💡 **Tip:** On Windows, extract to a simple path like `C:\vidya` — avoid folders with odd characters. Spaces are OK if you use quotes.

---

## 🐳 Step 1 — Install Docker

Docker is the one tool that runs the **entire backend** (API + database + cache) automatically. You don't need to understand how.

### Windows
1. Go to **https://www.docker.com/products/docker-desktop/** and download **Docker Desktop for Windows**.
2. Run the installer. Accept defaults. If it asks about **WSL 2**, say yes (it will guide you; you may need to restart).
3. Launch **Docker Desktop** from the Start menu. Wait until the whale icon shows **"Engine running"** (bottom-left corner, green).

### Mac
1. Go to **https://www.docker.com/products/docker-desktop/** and download **Docker Desktop for Mac**.
2. Open the downloaded `.dmg`, drag the Docker whale into Applications, and launch it.
3. Accept the terms. Wait until the whale icon in the top menu bar is steady (not animating).

### Linux (Debian/Ubuntu)
```bash
sudo apt-get update
sudo apt-get install -y docker.io docker-compose-plugin
sudo systemctl enable --now docker
sudo usermod -aG docker $USER
```
Then **log out and log back in** (so the group change takes effect).

### Linux (Arch)
```bash
sudo pacman -S docker docker-compose
sudo systemctl enable --now docker
sudo usermod -aG docker $USER
```
Then **log out and log back in**.

### ✅ Verify Docker is installed
Open a terminal and run:

```bash
docker --version
docker compose version
```

Both should print version numbers. If `docker compose` says `unknown shorthand flag: 'f'`, the Compose plugin is missing (re-check the install step for your OS).

---

## 🚀 Step 2 — Start the backend

The backend includes a **demo school** with 2 students and a working login — everything is created automatically on the first start. You don't type in any data.

1. In the terminal, make sure you're inside the extracted folder (from Step 0).
2. Run this and press Enter:

```bash
docker compose -f deploy/docker-compose.yml up --build
```

3. **First time only:** this builds the backend (a few minutes — you'll see lots of scrolling text, that's normal). When it's done it keeps running and shows a line like:

```
[GIN-debug] ... 
msg="api listening" addr=":8080"
```

4. **Leave this terminal open** (don't close it — that's your backend, like a radio playing in the background).
5. Open a **second terminal window** (same way as Step 0) and verify it's alive:

```bash
curl -s http://localhost:8080/health/ready
```

You should see: `{"db":"up","status":"ready"}`

> 🍎 **Mac note:** If curl isn't installed, that's fine — skip the check and move on; the browser step will confirm everything works.
> 🪟 **Windows note:** If `curl` complains, type `curl.exe` instead.

---

## 📱 Step 3 — Install Flutter

Flutter is what builds and runs the app you'll see in the browser.

### Windows
1. Download the **Windows ZIP** from **https://docs.flutter.dev/get-started/install/windows**
   (the button says "Download the latest stable release").
2. Extract it **to `C:\src\flutter`** (create `C:\src` first if needed). Do **not** put it in Program Files.
3. Add Flutter to your PATH:
   - Press `Win`, search **"Environment Variables"**, open it.
   - Under *User variables*, select **Path** → **Edit** → **New** → paste `C:\src\flutter\bin` → OK → OK.
4. Enable **Developer Mode** (Flutter needs it):
   - Settings → Privacy & security → For developers → turn on **Developer Mode**.

### Mac
```bash
cd ~
curl -O https://storage.googleapis.com/flutter_infra_release/releases/stable/macos/flutter_macos_3.47.0-stable.zip
unzip flutter_macos_3.47.0-stable.zip
echo 'export PATH="$PATH:$HOME/flutter/bin"' >> ~/.zshrc
source ~/.zshrc
```
> If macOS complains the app is from an unidentified developer, run:
> `sudo xattr -dr com.apple.quarantine ~/flutter`

### Linux (Debian/Ubuntu)
```bash
cd ~
curl -O https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/flutter_linux_3.47.0-stable.tar.xz
tar xf flutter_linux_3.47.0-stable.tar.xz
echo 'export PATH="$PATH:$HOME/flutter/bin"' >> ~/.bashrc
source ~/.bashrc
```

### Linux (Arch)
```bash
sudo pacman -S flutter
```

### ✅ Verify Flutter
In the terminal (new window if you just installed it):

```bash
flutter --version
```

You should see something like `Flutter 3.47.0 • channel stable`. If it says `command not found`, the PATH step didn't take — re-read it, or open a brand-new terminal window.

---

## 🌐 Step 4 — Run the app in your browser

1. In the terminal, go into the `mobile` folder inside the extracted project:

```bash
cd mobile
```

2. One-time package download (a minute or two):

```bash
flutter pub get
```

3. **Important:** make sure the backend terminal (Step 2) is still running. Then launch the app:

```bash
flutter run -d web-server --web-port 8081 --dart-define=API_BASE_URL=http://localhost:8080/api/v1
```

4. Wait for it to compile (first time: 2–5 minutes, you'll see "Compiling..."). When it's ready it prints something like:

```
lib/main.dart is being served at http://localhost:8081
```

5. Open your browser and go to: **http://localhost:8081**

---

## 🎉 Step 5 — See it working

You should now see the **Vidya login screen** — green school icon, "Vidya", a sign-in form.

**Log in with the demo account:**

| Field | Value |
|-------|-------|
| Email or phone | `principal@greenwood.edu` |
| Password | `admin12345` |

Click **Sign in** → you land on the **Dashboard** showing live numbers:

- **Students: 2** (the demo school's students)
- **Teachers: 0**
- **Attendance today: —** (nobody marked attendance yet)
- **Pending fees: 0**
- **Collected: ₹0**

> The dashboard reads straight from the backend database. When you (or your friend) start adding students or fees through the API, these numbers change.

### Try the dark mode 🌙
If your computer is in dark mode, the app is dark too — the whole theme (colors, spacing, text) is driven by one theme file, so both look polished.

---

## 🔁 Everyday use (how to start it again tomorrow)

1. Start the backend (leave running): `docker compose -f deploy/docker-compose.yml up --build`
2. In another terminal: `cd mobile && flutter run -d web-server --web-port 8081 --dart-define=API_BASE_URL=http://localhost:8080/api/v1`
3. Open **http://localhost:8081**

**To stop everything:** press `Ctrl + C` in both terminals. Optionally, to wipe the database and start fresh:
`docker compose -f deploy/docker-compose.yml down -v` (then start again with `up --build`).

---

## 🧯 Common problems (and fixes)

| Symptom | Fix |
|---|---|
| `docker: unknown shorthand flag: 'f'` | The Compose plugin isn't installed — redo the Docker install step for your OS. |
| `docker compose up` hangs or port errors `5432` / `6379` already in use | Another database is already running on your PC. Stop it, or run `docker compose down` first. |
| `curl: connection refused` | The backend isn't ready yet — wait a few more seconds and check the first terminal for errors. |
| Login says `INVALID_CREDENTIALS` | The demo data didn't seed. Run `docker compose -f deploy/docker-compose.yml down -v` then `up --build` again. |
| `flutter: command not found` | PATH step missed — open a brand-new terminal, or re-check Step 3 for your OS. |
| Browser shows a blank/white page | First compile is slow — wait 2–5 minutes. If it stays blank, close the tab, restart `flutter run`, refresh. |
| `Cleartext HTTP ... not permitted` | Only happens on Android phones, not the browser. Ignore it — you're using the browser. |
| Everything works but numbers are all zero | That's correct! The demo school has 2 students and no fees yet. The numbers change as you add data. |

---

## 🔬 Optional: poke the backend directly (curious only)

With the backend running, you can ask it for data in the terminal. The responses are JSON — don't worry if they look messy.

```bash
# Ask the server if it's alive
curl -s http://localhost:8080/health/ready

# Get a login token (this is exactly what the app does when you sign in)
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"principal@greenwood.edu","password":"admin12345","device":{"device_id":"demo"}}'
```

The second command returns a big block of text including `"access_token":"..."` and `"roles":["school_admin"]` — that's your digital "key".

---

## 📦 For the person sharing this project (zip it correctly)

If you're zipping the project for a friend, **exclude** the junk so the zip stays small and clean:

```bash
cd /path/to/erp
zip -r vidya-friend.zip . \
  -x ".git/*" "backend/bin/*" "backend/.env" \
  "mobile/.dart_tool/*" "mobile/build/*" "mobile/.idea/*" "*.iml"
```

That keeps `deploy/` (needed for Docker), `mobile/`, `backend/`, and the docs. The friend then follows **Step 0 → Step 5** above.

> If the extracted zip's `mobile` folder is **missing the `web` folder**, run this once before `flutter run`:
> `cd mobile && flutter create . --platforms=web`
