import type React from "react"
import { useAuth } from "../contexts/AuthContext"

/**
 * Props for the Navbar component
 */
interface NavbarProps {
  /** Whether dark mode is enabled */
  darkMode: boolean
  /** Callback function when theme toggle button is clicked */
  onThemeToggle: () => void
}

/**
 * The navigation bar component that displays the app title, theme toggle, and authentication controls
 * @param props - The component props
 * @returns A navigation bar with theme and authentication controls
 */
export const Navbar: React.FC<NavbarProps> = ({ darkMode, onThemeToggle }) => {
  const { user, isAuthenticated, login, logout, loading } = useAuth()

  return (
    <div className="navbar bg-base-200 shadow-md">
      <div className="flex-1">
        <a href="/" className="btn btn-ghost text-xl">
          GoLink
        </a>
      </div>
      <div className="flex-none gap-2">
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

        {loading ? (
          <span
            className="loading loading-spinner loading-sm"
            role="status"
            aria-label="Loading"
          />
        ) : isAuthenticated ? (
          <div className="dropdown dropdown-end">
            <button
              className="btn btn-ghost btn-circle avatar"
              type="button"
              aria-label={`${user?.name}'s profile`}
            >
              <div className="w-10 rounded-full">
                {user?.picture ? (
                  <img
                    alt={`${user.name}'s profile`}
                    src={user.picture}
                    referrerPolicy="no-referrer"
                  />
                ) : (
                  <div className="bg-primary text-primary-content grid place-items-center">
                    {user?.name.substring(0, 2).toUpperCase()}
                  </div>
                )}
              </div>
            </button>
            <ul className="mt-3 z-[1] p-2 shadow menu menu-sm dropdown-content bg-base-100 rounded-box w-52">
              <li className="p-2 text-sm opacity-70">{user?.email}</li>
              <li>
                <button onClick={logout} className="text-error" type="button">
                  Logout
                </button>
              </li>
            </ul>
          </div>
        ) : (
          <button
            onClick={login}
            className="btn btn-sm btn-primary"
            type="button"
          >
            Login
          </button>
        )}
      </div>
    </div>
  )
}
