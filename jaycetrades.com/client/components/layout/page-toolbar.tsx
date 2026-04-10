"use client";

import type * as React from "react";

import { cn } from "@/lib/utils";

interface PageToolbarProps {
	title: string;
	subtitle?: string;
	primaryControls?: React.ReactNode;
	secondaryControls?: React.ReactNode;
	rightSlot?: React.ReactNode;
}

export function PageToolbar({
	title,
	subtitle,
	primaryControls,
	secondaryControls,
	rightSlot,
}: PageToolbarProps): React.JSX.Element {
	const hasControls = Boolean(primaryControls || secondaryControls || rightSlot);

	return (
		<div
			className={cn(
				"sticky top-0 z-10 border-b bg-background/80 backdrop-blur-md",
			)}
		>
			<div className="mx-auto flex max-w-[1200px] flex-col gap-3 px-4 py-4 sm:px-7 lg:flex-row lg:items-center lg:justify-between">
				<div className="min-w-0">
					<h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
					{subtitle && (
						<p className="mt-0.5 text-sm text-muted-foreground">{subtitle}</p>
					)}
				</div>
				{hasControls && (
					<div className="-mx-4 overflow-x-auto px-4 sm:mx-0 sm:px-0">
						<div className="flex flex-wrap items-center gap-3">
							{primaryControls}
							{secondaryControls}
							{rightSlot}
						</div>
					</div>
				)}
			</div>
		</div>
	);
}
