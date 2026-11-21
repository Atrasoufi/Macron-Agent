import { useStore } from "./store";

export type Lang = "en" | "fa";

const FA: Record<string, string> = {
  // Sidebar & Header
  "This PC": "این کامپیوتر",
  "Work on this computer": "کار روی همین سیستم",
  "Add a server…": "افزودن سرور…",
  "Open folder": "باز کردن پوشه",
  "Workspaces": "فضاهای کاری",
  "Home": "خانه",
  "General conversations": "گفتگوهای عمومی",
  "Settings": "تنظیمات",
  "Connecting…": "در حال اتصال…",
  "Engine ready": "موتور آماده است",
  "Reconnecting…": "تلاش برای اتصال مجدد…",
  "Search or run a command": "جستجو یا اجرای دستور",
  "New conversation": "گفتگوی جدید",
  "Threads": "گفتگوها",
  "Hide the plan": "بستن پنل",
  "Show changes": "نمایش تغییرات",
  "Update available": "آپدیت جدید موجود است",
  "Update now": "آپدیت کن",
  "Updating…": "در حال آپدیت…",
  "Updated — restart Macron": "آپدیت شد — مکرون را دوباره باز کنید",

  // Composer & Slash commands
  "Build": "ساخت",
  "Plan": "برنامه",
  "Changes": "تغییرات",
  "Start a new conversation": "شروع گفتگوی تازه",
  "Summarise the conversation to free room": "خلاصه‌سازی گفتگو برای آزاد کردن حافظه",
  "Pick the model": "انتخاب مدل هوش مصنوعی",
  "Plan first and change nothing": "ابتدا برنامه‌ریزی کن بدون تغییر فایل",
  "Let Macron make changes": "اجازه تغییر فایل و اجرای دستور",
  "Open another folder": "باز کردن پوشه دیگر",
  "Approvals, providers and theme": "تأییدیه‌ها، مدل‌ها و ظاهر",
  "Switch execution isolation mode": "تغییر حالت سندباکس و اجرا",
  "View git changes and diff": "مشاهده تغییرات Git و Diff",
  "Stop the current job": "توقف کار فعلی",
  "Attach files or images": "پیوست فایل یا عکس",
  "Drop files or images here to attach": "فایل یا عکس را اینجا رها کنید",
  "Uploading…": "در حال آپلود…",
  "No matching files": "فایلی پیدا نشد",
  "waiting": "در صف",
  "Add to what Macron is doing, or press Esc to stop": "ادامه دستور را بنویسید یا Esc را بزنید",
  "Describe the goal; Macron plans first and changes nothing": "هدف را توضیح دهید؛ مکرون ابتدا برنامه می‌چیند بدون تغییر فایل",
  "Ask Macron to build, fix, research or automate something": "از مکرون بخواهید بسازد، خطایابی کند یا خودکارسازی انجام دهد (@فایل یا Ctrl+V)",
  "Pick a model": "انتخاب مدل",

  // Plan & Changes Inspector
  "All steps done": "همه مراحل انجام شد",
  "Working": "در حال کار",
  "Planning": "در حال برنامه‌ریزی",
  "No git repository in this folder": "این پوشه مخزن Git نیست",
  "Working tree is clean": "تغییری در فایل‌ها ثبت نشده است",
  "Refresh": "به‌روزرسانی",
  "Revert": "بازگردانی",
  "Confirm revert?": "مطمئنید بازگردانی شود؟",
  "Cancel": "انصراف",
  "Select a file to inspect diff": "برای مشاهده Diff یک فایل را انتخاب کنید",
  "Loading diff…": "در حال بارگذاری Diff…",
  "No diff available": "تغییری برای نمایش نیست",

  // Feed & Intro
  "Connect a model to start": "برای شروع یک مدل متصل کنید",
  "Connect a model": "اتصال مدل",
  "Macron needs a language model: OpenAI, Anthropic, Gemini, OpenRouter, a local Ollama, or any compatible provider.":
    "مکرون به یک مدل زبانی نیاز دارد: OpenAI، Anthropic، Gemini، OpenRouter، Ollama محلی یا هر سرویس سازگار دیگر.",
  "Conversations that don't belong to a project live here. Open a folder to give Macron a project to work in.":
    "گفتگوهای عمومی اینجا ذخیره می‌شوند. برای کار روی پروژه، یک پوشه باز کنید.",
  "Macron reads and changes files in this folder, runs commands here, and looks things up when it needs to.":
    "مکرون فایل‌های این پوشه را می‌خواند و ویرایش می‌کند، دستورات را اجرا می‌کند و در صورت نیاز جستجو می‌کند.",
  "Command palette": "پالت دستورات",
  "Plan before changing anything": "برنامه‌ریزی قبل از تغییر",
  "Copy": "کپی",
  "Copied": "کپی شد",
  "Details": "جزئیات",

  // Settings & Sections
  "General": "عمومی",
  "Appearance": "ظاهر",
  "System": "سیستم",
  "Light": "روشن",
  "Dark": "تیره",
  "Language": "زبان رابط کاربری",
  "Updates": "به‌روزرسانی خودکار",
  "Check for updates": "بررسی آپدیت جدید",
  "Checking…": "در حال بررسی…",
  "Up to date": "آخرین نسخه نصب است",
  "Install update": "نصب آپدیت از گیت‌هاب",
  "Release notes": "تغییرات این نسخه",
  "Safety": "امنیت و کنترل",
  "Command approvals": "تأیید دستورات",
  "Smart": "هوشمند",
  "Always ask": "همیشه بپرس",
  "Never ask": "نپرس",
  "Execution mode": "حالت اجرای دستورات",
  "Sandbox": "سندباکس",
  "Best-effort": "خودکار",
  "Local": "محلی",
  "YOLO for 30 minutes": "حالت آزاد (YOLO) برای ۳۰ دقیقه",
  "Model": "مدل هوش مصنوعی",
  "Effort": "میزان دقت",
  "Context window": "پنجره کانتکست",
  "Temperature": "دما (Temperature)",
  "MCP servers and plugins": "سرورهای MCP و پلاگین‌ها",
  "Add an MCP server…": "افزودن سرور MCP…",
  "Servers": "سرورها (SSH / WSL)",
  "Skills": "مهارت‌ها (Skills)",
  "Filter skills…": "جستجو در مهارت‌ها…",
  "No skills loaded.": "مهارتی بارگذاری نشده است.",
  "Memory": "حافظه بلندمدت (Memory)",
  "Agent memory": "حافظه کاری ایجنت",
  "About you": "اطلاعات درباره شما",
  "Add a memory note…": "یادداشت جدید برای حافظه بنویسید…",
  "Add": "افزودن",
  "No saved memories yet.": "هنوز موردی در حافظه ذخیره نشده است.",
  "Connections": "اتصال سرویس‌ها",
  "Connected": "تنظیم شده",
  "Not set up": "تنظیم نشده",
  "Configure": "تنظیم",
  "Close": "بستن",
  "Save": "ذخیره",
  "Saving…": "در حال ذخیره…",
  "Keyboard": "میانبرهای صفحه‌کلید",
  "Send": "ارسال",
  "New line": "خط جدید",
  "Stop": "توقف",
  "Plan mode": "حالت برنامه",
  "Loading…": "در حال بارگذاری…",
  "Connect": "اتصال",
  "Disconnect": "قطع اتصال",
  "Remove": "حذف",
  "Dismiss": "بستن",
  "Safe commands run on their own. Anything risky waits for you.": "دستورات امن خودکار اجرا می‌شوند؛ کارهای پرخطر منتظر تأیید شما می‌مانند.",
  "Every command waits for your OK.": "هر دستوری قبل از اجرا منتظر تأیید شما می‌ماند.",
  "Nothing waits. Only for machines you can afford to break.": "هیچ دستوری منتظر تأیید نمی‌ماند (فقط برای محیط‌های آزمایشی).",
  "Strict bubblewrap isolation is active (fail-closed).": "سندباکس سخت‌گیرانه bubblewrap فعال است (Fail-Closed).",
  "Strict bubblewrap isolation (fail-closed if unavailable).": "سندباکس سخت‌گیرانه bubblewrap (در صورت نبودن متوقف می‌شود).",
  "Strict bubblewrap isolation is only available on Linux.": "سندباکس bubblewrap فقط روی لینوکس در دسترس است.",
  "Sandbox active; falls back to local if unavailable.": "سندباکس فعال است؛ در صورت نبودن، محلی اجرا می‌شود.",
  "Unsandboxed; uses bubblewrap when available.": "بدون سندباکس؛ در صورت وجود bubblewrap از آن استفاده می‌شود.",
  "Unsandboxed (bubblewrap sandbox is Linux-only).": "بدون سندباکس (bubblewrap فقط مخصوص لینوکس است).",
  "Sandbox bypassed. Commands run directly on the host.": "سندباکس خاموش است و دستورات مستقیم روی سیستم اجرا می‌شوند.",
  "On. Commands run without asking until it runs out.": "روشن است. دستورات تا پایان زمان بدون پرسیدن اجرا می‌شوند.",
  "Let Macron run everything without asking, for a while.": "به مکرون اجازه دهید برای ۳۰ دقیقه همه‌چیز را بدون پرسیدن اجرا کند.",
};

export function tr(lang: Lang, key: string): string {
  if (lang === "fa") return FA[key] || key;
  return key;
}

export function useT() {
  const lang = useStore((s) => s.lang);
  return (key: string) => tr(lang, key);
}

export function applyLang(lang: Lang) {
  const root = document.documentElement;
  root.lang = lang;
  root.dir = lang === "fa" ? "rtl" : "ltr";
}
