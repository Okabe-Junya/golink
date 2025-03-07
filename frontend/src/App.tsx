import type React from "react"
import { useState, useEffect, useCallback } from "react"
import axios, { type AxiosError } from "axios"
import type { Link } from "./types/link"
import { Navbar } from "./components/Navbar"
import { LinkForm } from "./components/LinkForm"
import { LinkList } from "./components/LinkList"

const API_BASE_URL = "http://localhost:8080/api"
const APP_DOMAIN = process.env.REACT_APP_DOMAIN || "example.com"

const App: React.FC = () => {
  const [url, setUrl] = useState<string>("")
  const [short, setShort] = useState<string>("")
  const [links, setLinks] = useState<Link[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [editMode, setEditMode] = useState<boolean>(false)
  const [editingLink, setEditingLink] = useState<Link | null>(null)
  const [darkMode, setDarkMode] = useState<boolean>(false)

  useEffect(() => {
    const prefersDark = window.matchMedia(
      "(prefers-color-scheme: dark)",
    ).matches
    setDarkMode(prefersDark)
    document.documentElement.setAttribute(
      "data-theme",
      prefersDark ? "dark" : "light",
    )
  }, [])

  const fetchLinks = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await axios.get<Link[]>(`${API_BASE_URL}/links`)
      setLinks(res.data || [])
    } catch (error) {
      console.error("Error fetching links:", error)
      setError("Failed to fetch links. Please try again.")
      setLinks([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchLinks()
  }, [fetchLinks])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    setSuccess(null)

    try {
      const linkData = {
        short,
        url,
        access_level: "Public",
        allowed_users: [],
      }
      await axios.post(`${API_BASE_URL}/links`, linkData)
      resetForm()
      setSuccess("Link created successfully!")
      fetchLinks()
    } catch (error: unknown) {
      console.error("Error creating link:", error)
      const axiosError = error as AxiosError
      setError(
        (axiosError.response?.data as string) ||
          "Failed to create link. Please try again.",
      )
    } finally {
      setLoading(false)
    }
  }

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editingLink) return

    setLoading(true)
    setError(null)
    setSuccess(null)

    try {
      const linkData = {
        url,
      }
      await axios.put(`${API_BASE_URL}/links/${editingLink.short}`, linkData)
      setEditMode(false)
      setEditingLink(null)
      resetForm()
      setSuccess("Link updated successfully!")
      fetchLinks()
    } catch (error: unknown) {
      console.error("Error updating link:", error)
      const axiosError = error as AxiosError
      setError(
        (axiosError.response?.data as string) ||
          "Failed to update link. Please try again.",
      )
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (shortCode: string) => {
    if (!window.confirm("Are you sure you want to delete this link?")) return

    setLoading(true)
    setError(null)
    setSuccess(null)

    try {
      await axios.delete(`${API_BASE_URL}/links/${shortCode}`)
      setSuccess("Link deleted successfully!")
      fetchLinks()
    } catch (error: unknown) {
      console.error("Error deleting link:", error)
      const axiosError = error as AxiosError
      setError(
        (axiosError.response?.data as string) ||
          "Failed to delete link. Please try again.",
      )
    } finally {
      setLoading(false)
    }
  }

  const enterEditMode = (link: Link) => {
    setEditMode(true)
    setEditingLink(link)
    setUrl(link.url)
    setShort(link.short)
  }

  const resetForm = () => {
    setUrl("")
    setShort("")
    setEditMode(false)
    setEditingLink(null)
  }

  const handleCopy = (shortCode: string) => {
    const fullUrl = `http://${APP_DOMAIN}/${shortCode}`
    navigator.clipboard.writeText(fullUrl)
    setSuccess(`Copied ${fullUrl} to clipboard!`)
    setTimeout(() => setSuccess(null), 3000)
  }

  const toggleTheme = () => {
    const newTheme = darkMode ? "light" : "dark"
    setDarkMode(!darkMode)
    document.documentElement.setAttribute("data-theme", newTheme)
  }

  return (
    <div className="min-h-screen bg-base-100">
      <Navbar darkMode={darkMode} onThemeToggle={toggleTheme} />

      <div className="container mx-auto px-4 py-8">
        {error && (
          <div className="alert alert-error mb-4">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="stroke-current shrink-0 h-6 w-6"
              fill="none"
              viewBox="0 0 24 24"
              role="img"
              aria-label="Error icon"
            >
              <title>Error</title>
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>{error}</span>
            <button
              type="button"
              className="btn btn-sm btn-ghost"
              onClick={() => setError(null)}
            >
              Dismiss
            </button>
          </div>
        )}

        {success && (
          <div className="alert alert-success mb-4">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="stroke-current shrink-0 h-6 w-6"
              fill="none"
              viewBox="0 0 24 24"
              role="img"
              aria-label="Success icon"
            >
              <title>Success</title>
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>{success}</span>
            <button
              type="button"
              className="btn btn-sm btn-ghost"
              onClick={() => setSuccess(null)}
            >
              Dismiss
            </button>
          </div>
        )}

        <LinkForm
          url={url}
          short={short}
          editMode={editMode}
          loading={loading}
          onUrlChange={setUrl}
          onShortChange={setShort}
          onSubmit={editMode ? handleUpdate : handleSubmit}
          onCancel={resetForm}
          appDomain={APP_DOMAIN}
        />

        <LinkList
          links={links}
          loading={loading}
          appDomain={APP_DOMAIN}
          onEdit={enterEditMode}
          onDelete={handleDelete}
          onCopy={handleCopy}
        />
      </div>

      <footer className="footer footer-center p-4 bg-base-200 text-base-content mt-8">
        <div>
          <p>
            © 2025 - <a href="https://okabe-junya.github.io/">Junya Okabe</a>
          </p>
        </div>
      </footer>
    </div>
  )
}

export default App
