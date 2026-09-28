import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import * as api from "../api";
import { CoverArt } from "../components/CoverArt";
import type { AuthorDetail, Book } from "../types";

export function AuthorDetailPage() {
  const { id } = useParams();
  const [author, setAuthor] = useState<AuthorDetail | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    api
      .getAuthor(Number(id))
      .then(setAuthor)
      .catch((e: Error) => setError(e.message));
  }, [id]);

  if (error) return <p className="p-6 text-accent">{error}</p>;
  if (!author) return <p className="p-6 text-muted">Loading…</p>;

  return (
    <div className="mx-auto max-w-4xl px-6 py-8">
      <Link to="/authors" className="text-xs text-muted no-underline">
        ← Authors
      </Link>
      <h1 className="mt-2 font-serif text-3xl text-heading">{author.name}</h1>
      <p className="mt-1 text-sm text-muted">
        {author.bookCount} books · {author.readCount} read · {author.unreadOwned} unread owned ·{" "}
        {author.unreadCoveting} coveting
      </p>
      <BookRow title="Owned" books={author.owned} />
      <BookRow title="Coveting" books={author.coveting} />
    </div>
  );
}

function BookRow({ title, books }: { title: string; books: Book[] }) {
  if (books.length === 0) return null;
  return (
    <section className="mt-8">
      <h2 className="text-xs font-medium uppercase tracking-widest text-muted">{title}</h2>
      <div className="mt-3 grid grid-cols-2 gap-3 sm:grid-cols-4 md:grid-cols-6">
        {books.map((b) => (
          <Link key={b.id} to={`/?q=${encodeURIComponent(b.title)}`} className="no-underline">
            <div className="aspect-[2/3] overflow-hidden rounded-md bg-surface">
              <CoverArt title={b.title} coverUrl={b.coverUrl} />
            </div>
            <div className="mt-1 truncate text-xs text-heading">{b.title}</div>
          </Link>
        ))}
      </div>
    </section>
  );
}
