import type * as React from "react";

import { cn } from "@/lib/utils";

interface MetricProps {
	label: string;
	value: string | React.ReactNode;
	className?: string;
	align?: "left" | "right";
}

export function Metric({
	label,
	value,
	className,
	align = "left",
}: MetricProps): React.JSX.Element {
	return (
		<div
			className={cn("flex items-baseline justify-between gap-3", className)}
		>
			<span className="text-xs text-muted-foreground">{label}</span>
			{typeof value === "string" ? (
				<span
					className={cn(
						"text-sm tabular-nums",
						align === "right" ? "font-semibold" : "font-medium",
					)}
				>
					{value}
				</span>
			) : (
				value
			)}
		</div>
	);
}
