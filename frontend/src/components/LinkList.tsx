import type React from "react"
import type { Link } from "../types/link"
import { formatDate, formatDateWithTime } from "../utils/date"

/**
 * Props for the LinkList component
 */
interface LinkListProps {
  /** Array of links to display in the list */
  links: Link[]
  /** Whether the component is loading data */
  loading: boolean
  /** The domain where shortened links are hosted */
  appDomain: string
  /** Callback function when edit button is clicked */
  onEdit: (link: Link) => void
  /** Callback function when delete button is clicked */
  onDelete: (shortCode: string) => void
  /** Callback function when copy button is clicked */
  onCopy: (shortCode: string) => void
  /** Callback function when a link is selected for analytics */
  onViewAnalytics?: (linkId: string) => void
  /** Callback function when generate QR code button is clicked */
  onGenerateQrCode?: (linkUrl: string) => void
  /** Callback function for bulk deleting expired links */
  onBulkDeleteExpired?: () => void
  /** Whether there are any expired links */
  hasExpiredLinks?: boolean
}

/**
 * A component that displays a table of shortened links with actions
 */
export const LinkList: React.FC<LinkListProps> = ({
  links,
  loading,
  appDomain,
  onEdit,
  onDelete,
  onCopy,
  onViewAnalytics,
  onGenerateQrCode,
  onBulkDeleteExpired,
  hasExpiredLinks,
}) => {
  return (
    <div className="card bg-base-200 shadow-xl">
      <div className="card-body">
        <div className="flex justify-between items-center mb-4">
          <h2 className="card-title text-2xl">All Links</h2>
          {hasExpiredLinks && onBulkDeleteExpired && (
            <button
              type="button"
              onClick={onBulkDeleteExpired}
              className="btn btn-error btn-sm"
              title="Delete all expired links"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-4 w-4 mr-2"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                role="img"
                aria-label="Delete expired links icon"
              >
                <title>Delete expired links</title>
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                />
              </svg>
              Delete Expired Links
            </button>
          )}
        </div>
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
                <th>Expires</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {links && links.length === 0 && !loading ? (
                <tr>
                  <td colSpan={7} className="text-center py-4">
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
                  <tr
                    key={link.short}
                    className={`${link.is_expired ? "opacity-50" : ""}`}
                  >
                    <td>
                      <div className="flex items-center space-x-2">
                        <button
                          type="button"
                          className="btn btn-xs btn-ghost"
                          onClick={() => onCopy(link.short)}
                          aria-label="Copy to clipboard"
                          disabled={link.is_expired}
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
                      <div className="flex flex-col gap-1">
                        <div className="badge badge-outline">
                          {link.access_level}
                        </div>
                        {link.access_level === "restricted" &&
                          link.allowed_users.length > 0 && (
                            <div className="text-xs opacity-70">
                              Users: {link.allowed_users.join(", ")}
                            </div>
                          )}
                      </div>
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
                      {link.expires_at ? (
                        <div
                          className={`tooltip ${
                            link.is_expired
                              ? "text-error"
                              : new Date(link.expires_at) <
                                  new Date(Date.now() + 24 * 60 * 60 * 1000)
                                ? "text-warning"
                                : new Date(link.expires_at) <
                                    new Date(
                                      Date.now() + 7 * 24 * 60 * 60 * 1000,
                                    )
                                  ? "text-info"
                                  : ""
                          }`}
                          data-tip={formatDateWithTime(link.expires_at)}
                        >
                          {formatDate(link.expires_at)}
                          {link.is_expired && (
                            <span className="ml-1 badge badge-sm badge-error">
                              Expired
                            </span>
                          )}
                          {!link.is_expired &&
                            new Date(link.expires_at) <
                              new Date(Date.now() + 24 * 60 * 60 * 1000) && (
                              <span className="ml-1 badge badge-sm badge-warning">
                                Expires today
                              </span>
                            )}
                          {!link.is_expired &&
                            new Date(link.expires_at) >
                              new Date(Date.now() + 24 * 60 * 60 * 1000) &&
                            new Date(link.expires_at) <
                              new Date(
                                Date.now() + 7 * 24 * 60 * 60 * 1000,
                              ) && (
                              <span className="ml-1 badge badge-sm badge-info">
                                Expires soon
                              </span>
                            )}
                        </div>
                      ) : (
                        <span className="text-sm opacity-70">Never</span>
                      )}
                    </td>
                    <td>
                      <div className="flex space-x-1">
                        {onViewAnalytics && (
                          <button
                            type="button"
                            onClick={() => onViewAnalytics(link.short)}
                            className="btn btn-xs btn-primary"
                            aria-label="View analytics"
                          >
                            <svg
                              xmlns="http://www.w3.org/2000/svg"
                              className="h-4 w-4"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke="currentColor"
                              role="img"
                              aria-label="Analytics icon"
                            >
                              <title>View analytics</title>
                              <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                              />
                            </svg>
                          </button>
                        )}
                        {onGenerateQrCode && (
                          <button
                            type="button"
                            onClick={() =>
                              onGenerateQrCode(
                                `http://${appDomain}/${link.short}`,
                              )
                            }
                            className="btn btn-xs btn-secondary"
                            aria-label="Generate QR Code"
                          >
                            <svg
                              xmlns="http://www.w3.org/2000/svg"
                              className="h-4 w-4"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke="currentColor"
                              role="img"
                              aria-label="QR Code icon"
                            >
                              <title>Generate QR Code</title>
                              <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z"
                              />
                            </svg>
                          </button>
                        )}
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
