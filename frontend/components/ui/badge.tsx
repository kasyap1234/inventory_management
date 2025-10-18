import * as React from "react"
import { cn } from "@/lib/utils"

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'secondary' | 'success' | 'warning' | 'danger'
}

const Badge = React.forwardRef<HTMLDivElement, BadgeProps>(
  ({ className, variant = 'default', ...props }, ref) => {
    const variants = {
      default: "bg-blue-100 text-blue-700 border border-blue-200",
      secondary: "bg-gray-100 text-gray-700 border border-gray-200",
      success: "bg-emerald-100 text-emerald-700 border border-emerald-200",
      warning: "bg-amber-100 text-amber-700 border border-amber-200",
      danger: "bg-red-100 text-red-700 border border-red-200",
    }

    return (
      <div
        ref={ref}
        className={cn(
          "inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold transition-all duration-200 hover:scale-105",
          variants[variant],
          className
        )}
        {...props}
      />
    )
  }
)
Badge.displayName = "Badge"

export { Badge }
