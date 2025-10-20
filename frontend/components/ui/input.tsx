import * as React from "react"
import { cn } from "@/lib/utils"
import { AlertCircle } from "lucide-react"

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: string
  icon?: React.ReactNode
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, error, icon, ...props }, ref) => {
    return (
      <div className="relative w-full">
        {icon && (
          <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
            {icon}
          </div>
        )}
        <input
          type={type}
          className={cn(
            "flex h-11 w-full rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm transition-all",
            "placeholder:text-gray-400",
            "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/10 focus-visible:border-blue-400",
            "disabled:cursor-not-allowed disabled:opacity-50",
            "hover:border-gray-300",
            icon && "pl-10",
            error && "border-red-500 focus-visible:ring-red-500/20 focus-visible:border-red-500",
            className
          )}
          ref={ref}
          aria-invalid={error ? "true" : "false"}
          aria-describedby={error ? `${props.id}-error` : undefined}
          {...props}
        />
        {error && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2 text-red-500">
            <AlertCircle className="h-4 w-4" />
          </div>
        )}
        {error && (
          <p id={`${props.id}-error`} className="mt-1 text-sm text-red-600">
            {error}
          </p>
        )}
      </div>
    )
  }
)
Input.displayName = "Input"

export { Input }
