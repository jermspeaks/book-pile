import { NavLink, Outlet } from "react-router-dom";

const links = [
  { to: "/", label: "Library", end: true },
  { to: "/authors", label: "Authors" },
  { to: "/piles", label: "Piles" },
  { to: "/collections", label: "Collections" },
  { to: "/genres", label: "Genres" },
  { to: "/import", label: "Import" },
  { to: "/add", label: "Add" },
];

export function Layout() {
  return (
    <div className="flex h-full flex-col">
      <header className="flex shrink-0 items-center justify-between gap-6 border-b border-border px-6 py-3">
        <NavLink to="/" className="font-serif text-2xl text-heading no-underline">
          Book Pile
        </NavLink>
        <nav className="flex flex-wrap items-center gap-4 text-sm">
          {links.map((l) => (
            <NavLink
              key={l.to}
              to={l.to}
              end={l.end}
              className={({ isActive }) =>
                `tracking-wide no-underline ${isActive ? "text-heading" : "text-muted hover:text-heading"}`
              }
            >
              {l.label}
            </NavLink>
          ))}
        </nav>
      </header>
      <main className="min-h-0 flex-1">
        <Outlet />
      </main>
    </div>
  );
}
