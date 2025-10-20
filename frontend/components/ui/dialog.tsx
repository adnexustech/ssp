'use client'

import * as React from "react"

interface DialogProps {
  open?: boolean
  onOpenChange?: (open: boolean) => void
  children?: React.ReactNode
}

interface DialogTriggerProps {
  asChild?: boolean
  children?: React.ReactNode
}

interface DialogContentProps {
  className?: string
  children?: React.ReactNode
}

interface DialogHeaderProps {
  className?: string
  children?: React.ReactNode
}

interface DialogTitleProps {
  className?: string
  children?: React.ReactNode
}

interface DialogDescriptionProps {
  className?: string
  children?: React.ReactNode
}

const DialogContext = React.createContext<{
  open: boolean
  onOpenChange: (open: boolean) => void
}>({ open: false, onOpenChange: () => {} })

export function Dialog({ open = false, onOpenChange = () => {}, children }: DialogProps) {
  const [isOpen, setIsOpen] = React.useState(open)

  React.useEffect(() => {
    setIsOpen(open)
  }, [open])

  const handleOpenChange = (newOpen: boolean) => {
    setIsOpen(newOpen)
    onOpenChange(newOpen)
  }

  return (
    <DialogContext.Provider value={{ open: isOpen, onOpenChange: handleOpenChange }}>
      {children}
    </DialogContext.Provider>
  )
}

export function DialogTrigger({ asChild, children }: DialogTriggerProps) {
  const { onOpenChange } = React.useContext(DialogContext)

  if (asChild && React.isValidElement(children)) {
    return React.cloneElement(children, {
      onClick: () => onOpenChange(true),
    })
  }

  return (
    <button onClick={() => onOpenChange(true)}>
      {children}
    </button>
  )
}

export function DialogContent({ className, children }: DialogContentProps) {
  const { open, onOpenChange } = React.useContext(DialogContext)

  if (!open) return null

  return (
    <>
      <div
        className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm"
        onClick={() => onOpenChange(false)}
      />
      <div className="fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg sm:rounded-lg">
        <button
          className="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none"
          onClick={() => onOpenChange(false)}
        >
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4">
            <path d="M18 6 6 18" />
            <path d="m6 6 12 12" />
          </svg>
          <span className="sr-only">Close</span>
        </button>
        <div className={className}>
          {children}
        </div>
      </div>
    </>
  )
}

export function DialogHeader({ className, children }: DialogHeaderProps) {
  return (
    <div className={`flex flex-col space-y-1.5 text-center sm:text-left ${className || ''}`}>
      {children}
    </div>
  )
}

export function DialogTitle({ className, children }: DialogTitleProps) {
  return (
    <h2 className={`text-lg font-semibold leading-none tracking-tight ${className || ''}`}>
      {children}
    </h2>
  )
}

export function DialogDescription({ className, children }: DialogDescriptionProps) {
  return (
    <p className={`text-sm text-muted-foreground ${className || ''}`}>
      {children}
    </p>
  )
}

interface DialogFooterProps {
  className?: string
  children?: React.ReactNode
}

export function DialogFooter({ className, children }: DialogFooterProps) {
  return (
    <div className={`flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2 ${className || ''}`}>
      {children}
    </div>
  )
}