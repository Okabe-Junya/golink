import { render, screen, fireEvent, within } from "@testing-library/react"
import "@testing-library/jest-dom"
import { LinkList } from "../LinkList"
import type { Link } from "../../types/link"
import { vi } from "vitest"

describe("LinkList", () => {
  const mockLinks: Link[] = [
    {
      id: "1",
      url: "https://example.com",
      short: "test1",
      created_at: "2024-01-01T00:00:00Z",
      updated_at: "2024-01-01T00:00:00Z",
      created_by: "test-user",
      access_level: "public",
      allowed_users: [],
      click_count: 10,
      is_expired: false,
    },
    {
      id: "2",
      url: "https://test.com",
      short: "test2",
      created_at: "2024-01-02T00:00:00Z",
      updated_at: "2024-01-02T00:00:00Z",
      created_by: "test-user",
      access_level: "restricted",
      allowed_users: ["user1@example.com"],
      click_count: 5,
      is_expired: false,
    },
  ]

  const defaultProps = {
    links: mockLinks,
    loading: false,
    appDomain: "example.com",
    onEdit: vi.fn(),
    onDelete: vi.fn(),
    onCopy: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  test("renders list of links", () => {
    render(<LinkList {...defaultProps} />)
    expect(screen.getByText("test1")).toBeInTheDocument()
    expect(screen.getByText("test2")).toBeInTheDocument()
    expect(screen.getByText("https://example.com")).toBeInTheDocument()
    expect(screen.getByText("https://test.com")).toBeInTheDocument()
  })

  test("shows loading state", () => {
    render(<LinkList {...defaultProps} loading={true} />)
    expect(screen.getByRole("status")).toHaveClass("loading", "loading-spinner")
  })

  test("handles edit button click", () => {
    render(<LinkList {...defaultProps} />)
    const editButtons = screen.getAllByRole("button", { name: /Edit/i })
    fireEvent.click(editButtons[0])
    expect(defaultProps.onEdit).toHaveBeenCalledWith(mockLinks[0])
  })

  test("handles delete button click", () => {
    render(<LinkList {...defaultProps} />)
    const deleteButtons = screen.getAllByRole("button", { name: /Delete/i })
    fireEvent.click(deleteButtons[0])
    expect(defaultProps.onDelete).toHaveBeenCalledWith(mockLinks[0].short)
  })

  test("handles copy button click", () => {
    render(<LinkList {...defaultProps} />)
    const copyButtons = screen.getAllByRole("button", { name: /Copy/i })
    fireEvent.click(copyButtons[0])
    expect(defaultProps.onCopy).toHaveBeenCalledWith(mockLinks[0].short)
  })

  test("shows allowed users for restricted links", () => {
    render(<LinkList {...defaultProps} />)
    const row = screen.getByText("user1@example.com", { exact: false })
    expect(row).toBeInTheDocument()
  })

  test("links are properly formatted with domain", () => {
    render(<LinkList {...defaultProps} />)
    const shortLinks = screen.getAllByRole("link")
    expect(shortLinks[0]).toHaveAttribute("href", "http://example.com/test1")
  })

  test("external links open in new tab", () => {
    render(<LinkList {...defaultProps} />)
    const externalLinks = screen.getAllByRole("link")
    for (const link of externalLinks) {
      expect(link).toHaveAttribute("target", "_blank")
      expect(link).toHaveAttribute("rel", "noopener noreferrer")
    }
  })
})
