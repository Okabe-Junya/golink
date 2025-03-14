import { render, screen, fireEvent } from "@testing-library/react"
import "@testing-library/jest-dom"
import { Navbar } from "../Navbar"
import { vi } from "vitest"

// Mock the useAuth hook
const mockUseAuth = vi.fn()
vi.mock("../../contexts/AuthContext", () => ({
  useAuth: () => mockUseAuth(),
}))

describe("Navbar", () => {
  const defaultProps = {
    darkMode: false,
    onThemeToggle: vi.fn(),
  }

  beforeEach(() => {
    // Set default mock implementation
    mockUseAuth.mockReturnValue({
      user: {
        name: "Test User",
        email: "test@example.com",
        picture: "https://example.com/avatar.jpg",
      },
      isAuthenticated: true,
      login: vi.fn(),
      logout: vi.fn(),
      loading: false,
    })
  })

  test("renders navbar with title", () => {
    render(<Navbar {...defaultProps} />)
    expect(screen.getByText(/GoLink/i)).toBeInTheDocument()
  })

  test("handles theme toggle", () => {
    render(<Navbar {...defaultProps} />)
    const themeToggleButton = screen.getByRole("button", {
      name: /toggle theme/i,
    })
    fireEvent.click(themeToggleButton)
    expect(defaultProps.onThemeToggle).toHaveBeenCalled()
  })

  test("displays correct theme icon based on darkMode prop", () => {
    // Light mode
    const { rerender } = render(<Navbar {...defaultProps} />)
    expect(screen.getByRole("img", { name: /Dark mode/i })).toBeInTheDocument()

    // Dark mode
    rerender(<Navbar {...defaultProps} darkMode={true} />)
    expect(screen.getByRole("img", { name: /Light mode/i })).toBeInTheDocument()
  })

  test("displays user information when authenticated", () => {
    render(<Navbar {...defaultProps} />)
    const avatar = screen.getByRole("button", { name: "Test User's profile" })
    fireEvent.click(avatar)
    expect(screen.getByText("test@example.com")).toBeInTheDocument()
    expect(screen.getByRole("button", { name: /Logout/i })).toBeInTheDocument()
  })

  test("shows loading spinner when authentication is loading", () => {
    // Override the mock for this specific test
    mockUseAuth.mockReturnValueOnce({
      loading: true,
      isAuthenticated: false,
      login: vi.fn(),
      logout: vi.fn(),
      user: null,
    })

    render(<Navbar {...defaultProps} />)
    expect(screen.getByLabelText("Loading")).toBeInTheDocument()
  })
})
