import * as React from "react"
import { cn } from "@/lib/utils"
import { ChevronDown } from "lucide-react"

interface DropdownMenuProps {
  trigger: React.ReactNode
  children: React.ReactNode
  align?: "start" | "end"
  className?: string
}

export function DropdownMenu({ trigger, children, align = "end", className }: DropdownMenuProps) {
  const [isOpen, setIsOpen] = React.useState(false)
  const dropdownRef = React.useRef<HTMLDivElement>(null)

  React.useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside)
      return () => document.removeEventListener("mousedown", handleClickOutside)
    }
  }, [isOpen])

  return (
    <div className="relative inline-block" ref={dropdownRef}>
      <div onClick={() => setIsOpen(!isOpen)} className="cursor-pointer">
        {trigger}
      </div>
      
      {isOpen && (
        <div
          className={cn(
            "absolute z-50 mt-2 w-56 rounded-md border border-gray-200 bg-white shadow-lg",
            align === "end" ? "right-0" : "left-0",
            className
          )}
        >
          <div className="py-1">{children}</div>
        </div>
      )}
    </div>
  )
}

interface DropdownMenuItemProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode
  onSelect?: () => void
  disabled?: boolean
}

export function DropdownMenuItem({ 
  children, 
  onSelect, 
  disabled, 
  className,
  ...props 
}: DropdownMenuItemProps) {
  return (
    <div
      className={cn(
        "relative flex cursor-pointer select-none items-center px-3 py-2 text-sm outline-none transition-colors hover:bg-gray-100 focus:bg-gray-100",
        disabled && "pointer-events-none opacity-50",
        className
      )}
      onClick={() => {
        if (!disabled && onSelect) {
          onSelect()
        }
      }}
      {...props}
    >
      {children}
    </div>
  )
}

export function DropdownMenuSeparator() {
  return <div className="my-1 h-px bg-gray-200" />
}
