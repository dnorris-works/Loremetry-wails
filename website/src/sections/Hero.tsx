export default function Hero() {
  return (
    <section className="relative overflow-hidden bg-gradient-to-br from-slate-900 via-slate-800 to-indigo-900 text-white">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-indigo-500/20 via-transparent to-transparent" />
      <nav className="relative mx-auto flex max-w-6xl items-center justify-between px-6 py-6">
        <span className="text-xl font-bold tracking-tight">Loremetry</span>
        <div className="flex gap-6 text-sm font-medium text-slate-300">
          <a href="#features" className="hover:text-white transition">Features</a>
          <a href="#pricing" className="hover:text-white transition">Pricing</a>
        </div>
      </nav>

      <div className="relative mx-auto max-w-4xl px-6 pb-28 pt-20 text-center">
        <h1 className="text-5xl font-extrabold leading-tight tracking-tight sm:text-6xl">
          AI&#8209;powered manuscript analysis&nbsp;for&nbsp;writers
        </h1>
        <p className="mx-auto mt-6 max-w-2xl text-lg text-slate-300 leading-relaxed">
          Loremetry runs on your desktop, keeps your manuscripts private, and
          gives you detailed craft reports — from pacing and tension curves to
          show&#8209;don't&#8209;tell scoring — powered by AI.
        </p>
        <div className="mt-10 flex flex-col items-center gap-4 sm:flex-row sm:justify-center">
          <a
            href="#pricing"
            className="rounded-lg bg-indigo-500 px-8 py-3 text-base font-semibold shadow-lg shadow-indigo-500/30 transition hover:bg-indigo-400"
          >
            Download Free
          </a>
          <a
            href="#features"
            className="rounded-lg border border-slate-600 px-8 py-3 text-base font-semibold text-slate-200 transition hover:border-slate-400 hover:text-white"
          >
            Learn More
          </a>
        </div>
      </div>
    </section>
  );
}
