import { render, screen, fireEvent } from "@testing-library/react"
import "@testing-library/jest-dom"
import { LinkForm } from "../LinkForm"
import { vi } from "vitest"

describe("LinkForm", () => {
  const defaultProps = {
    url: "",
    short: "",
    editMode: false,
    loading: false,
    onUrlChange: vi.fn(),
    onShortChange: vi.fn(),
    onSubmit: vi.fn(),
    onCancel: vi.fn(),
    appDomain: "example.com",
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  test("renders form elements correctly", () => {
    render(<LinkForm {...defaultProps} />)
    expect(screen.getByLabelText(/Original URL/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Custom Short Code/i)).toBeInTheDocument()
    expect(screen.getByRole("button", { name: /Create/i })).toBeInTheDocument()
  })

  test("handles input changes", () => {
    render(<LinkForm {...defaultProps} />)
    const urlInput = screen.getByLabelText(/Original URL/i)
    const shortInput = screen.getByLabelText(/Custom Short Code/i)

    fireEvent.change(urlInput, { target: { value: "https://example.com" } })
    fireEvent.change(shortInput, { target: { value: "test" } })

    expect(defaultProps.onUrlChange).toHaveBeenCalledWith("https://example.com")
    expect(defaultProps.onShortChange).toHaveBeenCalledWith("test")
  })

  test("handles form submission", () => {
    const { container } = render(<LinkForm {...defaultProps} />)
    const form = container.querySelector("form")
    if (!form) throw new Error("Form not found")
    fireEvent.submit(form)
    expect(defaultProps.onSubmit).toHaveBeenCalled()
  })

  test("shows edit mode UI when editMode is true", () => {
    render(<LinkForm {...defaultProps} editMode={true} />)
    expect(screen.getByRole("button", { name: /Update/i })).toBeInTheDocument()
    expect(screen.getByRole("button", { name: /Cancel/i })).toBeInTheDocument()
  })

  test("disables form elements when loading", () => {
    render(<LinkForm {...defaultProps} loading={true} />)
    expect(screen.getByLabelText(/Original URL/i)).toBeDisabled()
    expect(screen.getByLabelText(/Custom Short Code/i)).toBeDisabled()
    expect(screen.getByRole("button", { name: /Create/i })).toBeDisabled()
  })

  test("validates URL format", () => {
    // まず無効なURLでレンダリング
    const { rerender } = render(<LinkForm {...defaultProps} url="not-a-url" />)
    const urlInput = screen.getByLabelText(/Original URL/i)

    // フォームをdirtyにする
    fireEvent.blur(urlInput)

    // 無効なURLではaria-invalidがtrueになる
    expect(urlInput).toHaveAttribute("aria-invalid", "true")

    // 有効なURLで再レンダリング
    rerender(<LinkForm {...defaultProps} url="https://example.com" />)

    // 有効なURLではaria-invalidがfalseになる
    expect(urlInput).toHaveAttribute("aria-invalid", "false")
    expect(urlInput).not.toHaveClass("input-error")
  })

  test("disables short code input in edit mode", () => {
    render(<LinkForm {...defaultProps} editMode={true} />)
    expect(screen.getByLabelText(/Custom Short Code/i)).toBeDisabled()
  })

  test("shows domain prefix for short code", () => {
    render(<LinkForm {...defaultProps} />)
    expect(screen.getByText("example.com/")).toBeInTheDocument()
  })

  test("cancels edit mode when cancel button is clicked", () => {
    render(<LinkForm {...defaultProps} editMode={true} />)
    fireEvent.click(screen.getByRole("button", { name: /Cancel/i }))
    expect(defaultProps.onCancel).toHaveBeenCalled()
  })
})
