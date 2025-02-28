import type React from "react"

interface NavbarProps {
  darkMode: boolean
  onThemeToggle: () => void
}

export const Navbar: React.FC<NavbarProps> = ({ darkMode, onThemeToggle }) => {
  return (
    <div className="navbar bg-base-200 shadow-md">
      <div className="flex-1">
        <a href="/" className="btn btn-ghost text-xl">
          GoLink
        </a>
      </div>
      <div className="flex-none">
        <button
          type="button"
          className="btn btn-ghost btn-circle"
          onClick={onThemeToggle}
          aria-label="toggle theme"
        >
          {darkMode ? (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              role="img"
              aria-label="Light mode"
            >
              <title>Light mode</title>
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
              />
            </svg>
          ) : (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              role="img"
              aria-label="Dark mode"
            >
              <title>Dark mode</title>
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
              />
            </svg>
          )}
        </button>
      </div>
    </div>
  )
}
