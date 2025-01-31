import type { ReactNode, PropsWithChildren } from "react";
import Link from "@docusaurus/Link";
import clsx from "clsx";

export function CardSection({
  id,
  title,
  children,
  description,
  className,
  hasSubSections = false,
  HeadingTag = "h3",
}: {
  id?: string;
  title: string;
  children: ReactNode;
  description?: ReactNode;
  hasSubSections?: boolean;
  HeadingTag?: keyof JSX.IntrinsicElements;
  className?: string;
}) {
  return (
    <div className={clsx("homepage-section", hasSubSections && "has-sub-sections", className)}>
      {title && <HeadingTag id={id ?? title}>{title}</HeadingTag>}
      {description && <p className="section-description">{description}</p>}
      <div className="section-content">{children}</div>
    </div>
  );
}

export function Card({
  id,
  date,
  title,
  description,
  to,
  tagLabel,
  className,
}: PropsWithChildren<{
  id?: string;
  date?: string;
  title: string;
  description?: string;
  to: string;
  tagLabel: string;
  className?: string;
}>) {
  const label = `Index page: ${title}`;
  return (
    <Link
      to={to}
      className={clsx("homepage-card", className)}
      data-goatcounter-click={to}
      data-goatcounter-title={label}
    >
      <div className="card-content">
        {date && <div className="description">{date}</div>}
        <div className="title" id={id && title}>
          {title}
        </div>
        {description && <div className="description">{description}</div>}
        {tagLabel && (
          <div className="tag-container">
            <span className="tag-label" style={{ backgroundColor: "inherit" }}>
              {tagLabel}
            </span>
          </div>
        )}
      </div>
    </Link>
  );
}
