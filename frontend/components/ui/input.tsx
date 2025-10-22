import * as React from "react"
import { cn } from "@/lib/utils"
import { AlertCircle, CheckCircle2, Loader2 } from "lucide-react"

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: string
  icon?: React.ReactNode
  isLoading?: boolean
  success?: boolean
  successMessage?: string
  helperText?: string
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, error, icon, isLoading, success, successMessage, helperText, ...props }, ref) => {
    const hasError = Boolean(error)
    const hasSuccess = Boolean(success && !hasError)
    const hasRightIcon = hasError || hasSuccess || isLoading

    return (
      <div className="relative w-full">
        {icon && (
          <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none">
            {icon}
          </div>
        )}
        <input
          type={type}
          className={cn(
            "flex h-12 w-full rounded-xl border border-gray-200 bg-white px-4 text-sm transition-all",
            "placeholder:text-gray-400",
            "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/10 focus-visible:border-blue-400",
            "disabled:cursor-not-allowed disabled:opacity-50 disabled:bg-gray-50",
            "hover:border-gray-300",
            "file:border-0 file:bg-transparent file:text-sm file:font-medium",
            icon && "pl-10",
            hasRightIcon && "pr-10",
            hasError && "border-red-300 focus-visible:ring-red-500/10 focus-visible:border-red-400 bg-red-50/30",
            hasSuccess && "border-emerald-300 focus-visible:ring-emerald-500/10 focus-visible:border-emerald-400 bg-emerald-50/30",
            className
          )}
          ref={ref}
          aria-invalid={hasError ? "true" : "false"}
          aria-describedby={error ? `${props.id}-error` : helperText ? `${props.id}-helper` : undefined}
          {...props}
        />
        {isLoading && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2 text-blue-500 pointer-events-none">
            <Loader2 className="h-4 w-4 animate-spin" />
          </div>
        )}
        {hasError && !isLoading && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2 text-red-500 pointer-events-none animate-fade-in">
            <AlertCircle className="h-4 w-4" />
          </div>
        )}
        {hasSuccess && !isLoading && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2 text-emerald-500 pointer-events-none animate-fade-in">
            <CheckCircle2 className="h-4 w-4" />
          </div>
        )}
        {error && (
          <p id={`${props.id}-error`} className="mt-1.5 text-sm text-red-600 animate-fade-in flex items-start gap-1">
            <span className="inline-block mt-0.5">{error}</span>
          </p>
        )}
        {!error && successMessage && (
          <p className="mt-1.5 text-sm text-emerald-600 animate-fade-in flex items-center gap-1">
            {successMessage}
          </p>
        )}
        {!error && !successMessage && helperText && (
          <p id={`${props.id}-helper`} className="mt-1.5 text-sm text-gray-500">
            {helperText}
          </p>
        )}
      </div>
    )
  }
)
Input.displayName = "Input"

export { Input }
