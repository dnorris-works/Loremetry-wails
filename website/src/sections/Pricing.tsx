const plans = [
  {
    name: "Free",
    price: "$0",
    period: "forever",
    description: "Download and run local analyses at no cost.",
    features: [
      "All local analyses included",
      "Line polish reports",
      "Print production checklists",
      "Vellum prep formatting",
      "Zeigarnik tension maps",
      "Unlimited projects",
    ],
    cta: "Download Free",
    href: "#",
    highlighted: false,
  },
  {
    name: "Writer",
    price: "Credits",
    period: "pay as you go",
    description: "Unlock AI-powered analyses with prepaid credits.",
    features: [
      "Everything in Free",
      "Show-don't-tell analysis",
      "Pacing & tension reports",
      "Dialogue quality scoring",
      "Character arc tracking",
      "More AI analyses added regularly",
    ],
    cta: "Get Started",
    href: "#",
    highlighted: true,
  },
];

export default function Pricing() {
  return (
    <section id="pricing" className="py-24">
      <div className="mx-auto max-w-5xl px-6">
        <h2 className="text-center text-3xl font-bold tracking-tight sm:text-4xl">
          Free to download. Pay only for AI.
        </h2>
        <p className="mx-auto mt-4 max-w-2xl text-center text-lg text-slate-600">
          Local analyses are always free. Add credits when you want
          AI&#8209;powered reports on your manuscript.
        </p>

        <div className="mt-16 grid gap-8 sm:grid-cols-2">
          {plans.map((plan) => (
            <div
              key={plan.name}
              className={`flex flex-col rounded-2xl border p-8 ${
                plan.highlighted
                  ? "border-indigo-500 bg-indigo-50 shadow-lg shadow-indigo-500/10"
                  : "border-slate-200 bg-white"
              }`}
            >
              <h3 className="text-lg font-semibold">{plan.name}</h3>
              <div className="mt-4 flex items-baseline gap-2">
                <span className="text-4xl font-extrabold tracking-tight">
                  {plan.price}
                </span>
                <span className="text-slate-500">/ {plan.period}</span>
              </div>
              <p className="mt-4 text-slate-600">{plan.description}</p>

              <ul className="mt-8 flex-1 space-y-3">
                {plan.features.map((f) => (
                  <li key={f} className="flex items-start gap-3 text-sm">
                    <span className="mt-0.5 text-indigo-500 font-bold">✓</span>
                    <span>{f}</span>
                  </li>
                ))}
              </ul>

              <a
                href={plan.href}
                className={`mt-8 block rounded-lg px-6 py-3 text-center font-semibold transition ${
                  plan.highlighted
                    ? "bg-indigo-500 text-white shadow-lg shadow-indigo-500/30 hover:bg-indigo-400"
                    : "border border-slate-300 text-slate-700 hover:border-slate-400"
                }`}
              >
                {plan.cta}
              </a>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
