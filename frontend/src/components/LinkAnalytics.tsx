import type React from "react"
import { useState, useEffect } from "react"
import axios from "axios"
import type { Link } from "../types/link"

interface LinkAnalyticsProps {
  linkId?: string
  apiBaseUrl: string
}

interface LinkStats {
  link_id: string
  short: string
  url: string
  click_count: number
  created_at: string
  age_days: number
  avg_clicks_per_day?: number
  access_level: string
  expires_at?: string
  is_expired?: boolean
}

const LinkAnalytics: React.FC<LinkAnalyticsProps> = ({
  linkId,
  apiBaseUrl,
}) => {
  const [linkStats, setLinkStats] = useState<LinkStats | null>(null)
  const [topLinks, setTopLinks] = useState<Link[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<"stats" | "top">("stats")

  useEffect(() => {
    if (activeTab === "stats" && linkId) {
      fetchLinkStats(linkId)
    } else if (activeTab === "top") {
      fetchTopLinks()
    }
  }, [linkId, activeTab])

  const fetchLinkStats = async (id: string) => {
    setLoading(true)
    setError(null)
    try {
      const response = await axios.get<LinkStats>(
        `${apiBaseUrl}/analytics/links/${id}`,
      )
      setLinkStats(response.data)
    } catch (err) {
      console.error("Error fetching link stats:", err)
      setError("Failed to load link statistics")
    } finally {
      setLoading(false)
    }
  }

  const fetchTopLinks = async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await axios.get<Link[]>(
        `${apiBaseUrl}/analytics/top?limit=10`,
      )
      setTopLinks(response.data)
    } catch (err) {
      console.error("Error fetching top links:", err)
      setError("Failed to load top links")
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center p-8">
        <div className="loading loading-spinner loading-lg text-primary" />
      </div>
    )
  }

  return (
    <div className="card bg-base-200 shadow-xl">
      <div className="card-body">
        <h2 className="card-title">Analytics</h2>

        <div className="tabs tabs-boxed mb-4">
          <button
            type="button"
            className={`tab ${activeTab === "stats" ? "tab-active" : ""}`}
            onClick={() => setActiveTab("stats")}
          >
            Link Statistics
          </button>
          <button
            type="button"
            className={`tab ${activeTab === "top" ? "tab-active" : ""}`}
            onClick={() => setActiveTab("top")}
          >
            Top Links
          </button>
        </div>

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
          </div>
        )}

        {activeTab === "stats" &&
          (linkId ? (
            <>
              {linkStats ? (
                <div className="overflow-x-auto">
                  <table className="table w-full">
                    <tbody>
                      <tr>
                        <td className="font-bold">Short Code</td>
                        <td>{linkStats.short}</td>
                      </tr>
                      <tr>
                        <td className="font-bold">Target URL</td>
                        <td className="break-all">
                          <a
                            href={linkStats.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="link link-primary"
                          >
                            {linkStats.url}
                          </a>
                        </td>
                      </tr>
                      <tr>
                        <td className="font-bold">Click Count</td>
                        <td>{linkStats.click_count}</td>
                      </tr>
                      <tr>
                        <td className="font-bold">Created</td>
                        <td>
                          {new Date(linkStats.created_at).toLocaleString()}
                        </td>
                      </tr>
                      <tr>
                        <td className="font-bold">Age</td>
                        <td>{linkStats.age_days.toFixed(1)} days</td>
                      </tr>
                      {linkStats.expires_at && (
                        <tr>
                          <td className="font-bold">Expires</td>
                          <td
                            className={`${linkStats.is_expired ? "text-error" : ""}`}
                          >
                            {new Date(linkStats.expires_at).toLocaleString()}
                            {linkStats.is_expired && (
                              <span className="ml-2 badge badge-sm badge-error">
                                Expired
                              </span>
                            )}
                          </td>
                        </tr>
                      )}
                      {linkStats.avg_clicks_per_day !== undefined && (
                        <tr>
                          <td className="font-bold">Avg. Clicks/Day</td>
                          <td>{linkStats.avg_clicks_per_day.toFixed(2)}</td>
                        </tr>
                      )}
                      <tr>
                        <td className="font-bold">Access Level</td>
                        <td>{linkStats.access_level}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="text-center py-4">
                  No statistics available for this link.
                </div>
              )}
            </>
          ) : (
            <div className="text-center py-4">
              Select a link to view its statistics.
            </div>
          ))}

        {activeTab === "top" &&
          (topLinks.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="table w-full">
                <thead>
                  <tr>
                    <th>Short Code</th>
                    <th>URL</th>
                    <th>Clicks</th>
                  </tr>
                </thead>
                <tbody>
                  {topLinks.map((link) => (
                    <tr key={link.id}>
                      <td>{link.short}</td>
                      <td className="truncate max-w-xs">
                        <a
                          href={link.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="link link-primary"
                        >
                          {link.url}
                        </a>
                      </td>
                      <td>{link.click_count}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-center py-4">
              No links available or links haven't been clicked yet.
            </div>
          ))}
      </div>
    </div>
  )
}

export default LinkAnalytics
