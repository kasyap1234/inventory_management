import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"
import { Loader2 } from "lucide-react"

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 active:scale-[0.98]",
  {
    variants: {
      variant: {
        default: "bg-gray-900 text-white hover:bg-gray-800 shadow-sm hover:shadow-md focus-visible:ring-gray-900/20",
        primary: "bg-blue-600 text-white hover:bg-blue-700 shadow-sm hover:shadow-md focus-visible:ring-blue-600/20",
        destructive: "bg-red-600 text-white hover:bg-red-700 shadow-sm hover:shadow-md focus-visible:ring-red-600/20",
        outline: "border border-gray-300 bg-white hover:bg-gray-50 text-gray-900 shadow-sm focus-visible:ring-gray-900/20",
        secondary: "bg-gray-100 text-gray-900 hover:bg-gray-200 shadow-sm focus-visible:ring-gray-900/20",
        ghost: "hover:bg-gray-100 text-gray-700 hover:text-gray-900",
        link: "text-blue-600 underline-offset-4 hover:underline hover:text-blue-700 p-0",
        success: "bg-green-600 text-white hover:bg-green-700 shadow-sm hover:shadow-md focus-visible:ring-green-600/20",
        warning: "bg-orange-600 text-white hover:bg-orange-700 shadow-sm hover:shadow-md focus-visible:ring-orange-600/20",
      },
      size: {
        default: "h-10 px-4 py-2 text-sm",
        sm: "h-9 px-3 py-1.5 text-xs",
        lg: "h-11 px-6 py-2.5 text-base",
        xl: "h-12 px-8 py-3 text-lg",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  isLoading?: boolean
  loadingText?: string
}

const Button = React.memo(
  React.forwardRef<HTMLButtonElement, ButtonProps>(
    ({ className, variant, size, isLoading, loadingText, children, disabled, ...props }, ref) => {
      return (
        <button
          className={cn(buttonVariants({ variant, size, className }))}
          ref={ref}
          disabled={disabled || isLoading}
          aria-busy={isLoading}
          {...props}
        >
          {isLoading && (
            <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
          )}
          <span className={cn(isLoading && "opacity-70")}>
            {isLoading && loadingText ? loadingText : children}
          </span>
        </button>
      )
    }
  )
)
Button.displayName = "Button"

export { Button, buttonVariants }
