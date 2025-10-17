import * as React from "react"
import { cn } from "@/lib/utils"

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'secondary' | 'success' | 'warning' | 'danger'
}

const Badge = React.forwardRef<HTMLDivElement, BadgeProps>(
  ({ className, variant = 'default', ...props }, ref) => {
    const variants = {
      default: "bg-gradient-to-r from-indigo-100 to-indigo-200 text-indigo-800 border border-indigo-200/50",
      secondary: "bg-gray-100 text-gray-800 border border-gray-200/50",
      success: "bg-gradient-to-r from-emerald-100 to-emerald-200 text-emerald-800 border border-emerald-200/50",
      warning: "bg-gradient-to-r from-amber-100 to-amber-200 text-amber-800 border border-amber-200/50",
      danger: "bg-gradient-to-r from-red-100 to-red-200 text-red-800 border border-red-200/50",
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
