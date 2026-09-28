import { coverHue } from "../lib/placeholder";

export function CoverArt({
  title,
  coverUrl,
  className = "",
}: {
  title: string;
  coverUrl?: string | null;
  className?: string;
}) {
  if (coverUrl) {
    return (
      <img
        src={coverUrl}
        alt=""
        className={`h-full w-full object-cover ${className}`}
      />
    );
  }
  const hue = coverHue(title);
  return (
    <div
      className={`flex h-full w-full items-end p-2 ${className}`}
      style={{
        background: `linear-gradient(165deg, hsl(${hue} 28% 42%), hsl(${(hue + 40) % 360} 22% 28%))`,
      }}
    >
      <span className="line-clamp-4 text-left text-xs leading-snug text-white/90">
        {title}
      </span>
    </div>
  );
}
