export default function Footer() {
  return (
    <footer className="border-t border-slate-200 bg-slate-50 py-12">
      <div className="mx-auto flex max-w-6xl flex-col items-center gap-4 px-6 text-sm text-slate-500 sm:flex-row sm:justify-between">
        <span>&copy; {new Date().getFullYear()} Loremetry. All rights reserved.</span>
        <div className="flex gap-6">
          <a href="#features" className="hover:text-slate-700 transition">Features</a>
          <a href="#pricing" className="hover:text-slate-700 transition">Pricing</a>
        </div>
      </div>
    </footer>
  );
}
