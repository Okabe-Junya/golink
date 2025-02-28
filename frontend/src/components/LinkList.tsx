import type React from "react"
import type { Link } from "../types/link"
import { formatDate, formatDateWithTime } from "../utils/date"

interface LinkListProps {
  links: Link[]
  loading: boolean
  appDomain: string
  onEdit: (link: Link) => void
  onDelete: (shortCode: string) => void
  onCopy: (shortCode: string) => void
}

export const LinkList: React.FC<LinkListProps> = ({
  links,
  loading,
  appDomain,
  onEdit,
  onDelete,
  onCopy,
}) => {
  return (
    <div className="card bg-base-200 shadow-xl">
      <div className="card-body">
        <h2 className="card-title text-2xl mb-4">All Links</h2>
        {loading && (
          <div className="flex justify-center my-4">
            <span
              className="loading loading-spinner loading-lg"
              role="status"
              aria-label="Loading"
            />
          </div>
        )}
        <div className="link-table-container">
          <table className="table table-zebra">
            <thead>
              <tr>
                <th>Short Code</th>
                <th>URL</th>
                <th>Access</th>
                <th>Clicks</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {links && links.length === 0 && !loading ? (
                <tr>
                  <td colSpan={6} className="text-center py-4">
                    <div className="flex flex-col items-center gap-2">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        className="h-10 w-10 text-base-content opacity-50"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        role="img"
                        aria-label="Empty state icon"
                      >
                        <title>No links found</title>
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M13 10V3L4 14h7v7l9-11h-7z"
                        />
                      </svg>
                      <p>No links found. Create your first link above!</p>
                    </div>
                  </td>
                </tr>
              ) : (
                links.map((link) => (
                  <tr key={link.short}>
                    <td>
                      <div className="flex items-center space-x-2">
                        <button
                          type="button"
                          className="btn btn-xs btn-ghost"
                          onClick={() => onCopy(link.short)}
                          aria-label="Copy to clipboard"
                        >
                          <svg
                            xmlns="http://www.w3.org/2000/svg"
                            className="h-4 w-4"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            role="img"
                            aria-label="Copy icon"
                          >
                            <title>Copy to clipboard</title>
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
                            />
                          </svg>
                        </button>
                        <a
                          href={`http://${appDomain}/${link.short}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="link link-hover"
                        >
                          {link.short}
                        </a>
                      </div>
                    </td>
                    <td>
                      <a
                        href={link.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="link link-hover truncate-url"
                        title={link.url}
                      >
                        {link.url}
                      </a>
                    </td>
                    <td>
                      <div className="badge badge-outline">
                        {link.access_level}
                      </div>
                      {link.access_level === "Restricted" &&
                        link.allowed_users.length > 0 && (
                          <div className="text-xs opacity-70 mt-1">
                            Users: {link.allowed_users.join(", ")}
                          </div>
                        )}
                    </td>
                    <td>
                      <div className="badge badge-neutral">
                        {link.click_count}
                      </div>
                    </td>
                    <td>
                      <div
                        className="tooltip"
                        data-tip={formatDateWithTime(link.created_at)}
                      >
                        {formatDate(link.created_at)}
                      </div>
                    </td>
                    <td>
                      <div className="flex space-x-1">
                        <button
                          type="button"
                          onClick={() => onEdit(link)}
                          className="btn btn-xs btn-info"
                          aria-label="Edit link"
                        >
                          <svg
                            xmlns="http://www.w3.org/2000/svg"
                            className="h-4 w-4"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            role="img"
                            aria-label="Edit icon"
                          >
                            <title>Edit link</title>
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                            />
                          </svg>
                        </button>
                        <button
                          type="button"
                          onClick={() => onDelete(link.short)}
                          className="btn btn-xs btn-error"
                          aria-label="Delete link"
                        >
                          <svg
                            xmlns="http://www.w3.org/2000/svg"
                            className="h-4 w-4"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            role="img"
                            aria-label="Delete icon"
                          >
                            <title>Delete link</title>
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                            />
                          </svg>
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
