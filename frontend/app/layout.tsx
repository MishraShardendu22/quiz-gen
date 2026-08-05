import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = {
  title: "QuizGen - AI Quiz Generator Dashboard",
  description: "AI-powered quiz generation platform and session tracker",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="h-full dark">
      <body className="min-h-full flex flex-col font-sans bg-[#090d16] text-slate-100 antialiased selection:bg-indigo-500 selection:text-white">
        <header className="bg-slate-900/80 border-b border-slate-800/80 sticky top-0 z-50 backdrop-blur-xl">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between h-16 items-center">
              <div className="flex items-center space-x-10">
                <Link href="/" className="flex items-center space-x-2">
                  <span className="text-2xl font-black text-transparent bg-clip-text bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 tracking-tight">
                    QuizGen
                  </span>
                  <span className="text-[10px] uppercase font-bold tracking-widest bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 px-2 py-0.5 rounded-full">
                    AI Studio
                  </span>
                </Link>

                <nav className="flex space-x-2">
                  <Link
                    href="/"
                    className="px-3.5 py-2 rounded-lg text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
                  >
                    Dashboard
                  </Link>
                  <Link
                    href="/topics"
                    className="px-3.5 py-2 rounded-lg text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
                  >
                    Topics
                  </Link>
                  <Link
                    href="/sessions"
                    className="px-3.5 py-2 rounded-lg text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
                  >
                    Sessions
                  </Link>
                  <Link
                    href="/usage"
                    className="px-3.5 py-2 rounded-lg text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
                  >
                    Usage & Costs
                  </Link>
                </nav>
              </div>

              <div className="flex items-center space-x-3">
                <Link
                  href="/topics"
                  className="px-4 py-2 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-lg text-xs font-bold shadow-lg shadow-indigo-600/20 hover:from-indigo-500 hover:to-purple-500 transition-all duration-200"
                >
                  + Generate Quiz
                </Link>
              </div>
            </div>
          </div>
        </header>

        <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {children}
        </main>
      </body>
    </html>
  );
}
