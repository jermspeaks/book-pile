export function StarRating({
  value,
  onChange,
}: {
  value?: number | null;
  onChange: (n: number) => void;
}) {
  const current = value ?? 0;
  return (
    <div className="flex gap-1" role="group" aria-label="Want to read rating">
      {[1, 2, 3, 4, 5].map((n) => (
        <button
          key={n}
          type="button"
          aria-label={`Want ${n}`}
          aria-pressed={current === n}
          className={`border-0 bg-transparent text-lg ${n <= current ? "text-accent" : "text-dim"}`}
          onClick={() => onChange(n)}
        >
          {n <= current ? "★" : "☆"}
        </button>
      ))}
    </div>
  );
}
