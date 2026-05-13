/**
 * Pawnix logo — paw print as a wifi-style signal beacon.
 *
 * Renders inline SVG so the panel and the signal both follow the active
 * theme (Tailwind `dark:` classes), rather than the OS-level
 * `prefers-color-scheme` setting. That way the logo flips together with
 * everything else when the user toggles the theme switch in the sidebar.
 *
 * The static `web/public/logo.svg` and `web/src/app/icon.svg` files use
 * a `prefers-color-scheme` media query for contexts we can't style with
 * Tailwind (favicon in the browser tab, README on GitHub).
 */
type LogoProps = {
  className?: string;
};

export function Logo({ className }: LogoProps) {
  return (
    <svg
      viewBox="0 0 100 100"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      role="img"
      aria-label="Pawnix logo"
    >
      {/* Panel: deep navy in dark mode; transparent in light mode so the
          logo sits flush against the page background instead of stamping
          a gray square onto a white surface. */}
      <rect
        width="100"
        height="100"
        rx="22"
        className="fill-transparent dark:fill-[#0B1120]"
      />

      {/* Outer signal ring */}
      <circle
        cx="50"
        cy="50"
        r="32"
        fill="none"
        strokeWidth="5"
        className="stroke-teal-600 dark:stroke-[#5EEAD4]"
      />

      {/* Wifi-style signal arcs */}
      <path
        d="M 38 34 Q 50 26 62 34"
        fill="none"
        strokeWidth="5"
        strokeLinecap="round"
        className="stroke-teal-600 dark:stroke-[#5EEAD4]"
      />
      <path
        d="M 30 48 Q 50 38 70 48"
        fill="none"
        strokeWidth="5"
        strokeLinecap="round"
        className="stroke-teal-600 dark:stroke-[#5EEAD4]"
      />

      {/* Source dot */}
      <circle
        cx="50"
        cy="60"
        r="6"
        className="fill-teal-600 dark:fill-[#5EEAD4]"
      />
    </svg>
  );
}
