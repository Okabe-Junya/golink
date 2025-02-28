import type React from "react"
import { Link } from "../types/link"

interface LinkFormProps {
  url: string
  short: string
  editMode: boolean
  loading: boolean
  onUrlChange: (value: string) => void
  onShortChange: (value: string) => void
  onSubmit: (e: React.FormEvent) => Promise<void>
  onCancel: () => void
  appDomain: string
}

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
  return (
    <div className="card bg-base-200 shadow-xl mb-8">
      <div className="card-body">
        <h2 className="card-title text-2xl mb-4">
          {editMode ? "Update Link" : "Create New Link"}
        </h2>
        <form onSubmit={onSubmit}>
          <div className="form-control mb-4">
            <label htmlFor="url" className="label">
              <span className="label-text">Original URL</span>
            </label>
            <input
              id="url"
              type="url"
              placeholder="https://example.com/very/long/url"
              className="input input-bordered w-full"
              value={url}
              onChange={(e) => onUrlChange(e.target.value)}
              required
              disabled={loading}
            />
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
              />
            </div>
            <span id="shortCodeHint" className="label-text-alt mt-2">
              Use a memorable word or phrase
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
              >
                Cancel
              </button>
            )}
            <button
              type="submit"
              className={`btn btn-primary ${loading ? "loading" : ""}`}
              disabled={loading}
            >
              {editMode ? "Update Link" : "Create Link"}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
