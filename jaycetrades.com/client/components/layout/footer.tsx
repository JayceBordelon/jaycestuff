import Link from "next/link";

export function Footer() {
	return (
		<div className="flex flex-wrap items-center justify-center gap-2 border-t px-7 py-3.5 text-[10px] text-muted-foreground">
			<span>
				<strong>Disclaimer:</strong> Not financial advice. Options
				trading involves substantial risk.
			</span>
			<span className="text-border">|</span>
			<Link
				href="/terms"
				className="underline underline-offset-2 transition-colors hover:text-foreground"
			>
				Terms
			</Link>
			<span className="text-border">|</span>
			<Link
				href="/faq"
				className="underline underline-offset-2 transition-colors hover:text-foreground"
			>
				FAQ
			</Link>
			<span className="text-border">|</span>
			<span>
				Built by{" "}
				<a
					href="https://jaycebordelon.com"
					target="_blank"
					rel="noopener noreferrer"
					className="underline underline-offset-2 transition-colors hover:text-foreground"
				>
					Jayce Bordelon
				</a>
			</span>
		</div>
	);
}
