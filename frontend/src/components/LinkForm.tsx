import type React from "react"
import { useState } from "react"

/**
 * Props for the LinkForm component
 */
interface LinkFormProps {
  /** The original URL to be shortened */
  url: string
  /** The custom short code for the URL */
  short: string
  /** Whether the form is in edit mode */
  editMode: boolean
  /** Whether the form is currently processing a request */
  loading: boolean
  /** Callback function when URL input changes */
  onUrlChange: (value: string) => void
  /** Callback function when short code input changes */
  onShortChange: (value: string) => void
  /** Callback function when form is submitted */
  onSubmit: (e: React.FormEvent) => Promise<void>
  /** Callback function when edit is cancelled */
  onCancel: () => void
  /** The domain where shortened links will be hosted */
  appDomain: string
}

/**
 * A form component for creating and editing shortened links
 */
export const LinkForm: React.FC<LinkFormProps> = ({
  url,
  short,
  editMode,
  loading,
  onUrlChange,
  onShortChange,
  onSubmit,
  onCancel,
  appDomain,
}) => {
  const [isDirty, setIsDirty] = useState(false)

  const validateUrl = (value: string): boolean => {
    if (!value) return true // Empty is valid (will be caught by required)
    try {
      new URL(value)
      return true
    } catch {
      return false
    }
  }

  const isUrlValid = validateUrl(url)
  const showError = isDirty && !isUrlValid && url !== ""

  const handleUrlChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setIsDirty(true)
    onUrlChange(e.target.value)
  }

  return (
    <div className="card bg-base-200 shadow-xl mb-8">
      <div className="card-body">
        <h2 className="card-title text-2xl mb-4">
          {editMode ? "Update Link" : "Create New Link"}
        </h2>
        <form onSubmit={onSubmit} noValidate>
          <div className="form-control mb-4">
            <label htmlFor="url" className="label">
              <span className="label-text">Original URL</span>
            </label>
            <div className="flex flex-col gap-1">
              <input
                id="url"
                type="url"
                placeholder="https://example.com/very/long/url"
                className={`input input-bordered w-full ${!isUrlValid && isDirty ? "input-error" : ""}`}
                value={url}
                onChange={handleUrlChange}
                onBlur={() => setIsDirty(true)}
                required
                disabled={loading}
                aria-invalid={!isUrlValid && isDirty}
              />
              {!isUrlValid && isDirty && url !== "" && (
                <div className="label">
                  <span className="label-text-alt text-error">
                    Please enter a valid URL
                  </span>
                </div>
              )}
            </div>
          </div>
          <div className="form-control mb-4">
            <label htmlFor="shortCode" className="label">
              <span className="label-text">Custom Short Code</span>
            </label>
            <div className="join w-full">
              <div className="join-item bg-base-300 px-3 flex items-center">
                {appDomain}/
              </div>
              <input
                id="shortCode"
                type="text"
                placeholder="foo"
                className="input input-bordered join-item w-full"
                value={short}
                onChange={(e) => onShortChange(e.target.value)}
                required
                disabled={editMode || loading}
                aria-describedby="shortCodeHint"
                pattern="[a-zA-Z0-9-_]+"
                title="Only letters, numbers, hyphens and underscores are allowed"
              />
            </div>
            <span id="shortCodeHint" className="label-text-alt mt-2">
              Use a memorable word or phrase (letters, numbers, hyphens and
              underscores only)
            </span>
          </div>
          <div className="form-control mb-4">
            <div className="badge badge-neutral">Access Level: Public</div>
            <div className="text-info text-sm mt-2">
              Access level configuration is not available in this version
            </div>
          </div>
          <div className="card-actions justify-end">
            {editMode && (
              <button
                type="button"
                className="btn btn-outline"
                onClick={onCancel}
                disabled={loading}
              >
                Cancel
              </button>
            )}
            <button
              type="submit"
              className={`btn btn-primary ${loading ? "loading" : ""}`}
              disabled={loading || showError}
            >
              {editMode ? "Update Link" : "Create Link"}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
