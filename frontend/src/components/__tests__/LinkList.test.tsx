import { render, screen, fireEvent } from "@testing-library/react"
import "@testing-library/jest-dom"
import { LinkList } from "../LinkList"
import type { Link } from "../../types/link"

describe("LinkList", () => {
  const mockLinks: Link[] = [
    {
      id: "1",
      url: "https://example.com",
      short: "test1",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      created_by: "test-user",
      access_level: "public",
      allowed_users: [],
      click_count: 0,
    },
    {
      id: "2",
      url: "https://test.com",
      short: "test2",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      created_by: "test-user",
      access_level: "public",
      allowed_users: [],
      click_count: 0,
    },
  ]

  const defaultProps = {
    links: mockLinks,
    loading: false,
    appDomain: "example.com",
    onEdit: jest.fn(),
    onDelete: jest.fn(),
    onCopy: jest.fn(),
  }

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

  test("displays empty state when no links", () => {
    render(<LinkList {...defaultProps} links={[]} />)
    expect(
      screen.getByText("No links found. Create your first link above!")
    ).toBeInTheDocument()
  })
})
