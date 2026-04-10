import type { Metadata } from "next";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export const metadata: Metadata = {
	title: "Terms of Service",
	description:
		"JayceTrades terms of service, risk disclosures, and legal disclaimers for AI-generated options trade suggestions.",
};

export default function TermsPage() {
	return (
		<div className="mx-auto max-w-3xl px-6 py-10">
			<h1 className="mb-2 text-2xl font-extrabold tracking-tight">
				Terms of Service
			</h1>
			<p className="mb-8 text-sm text-muted-foreground">
				Last updated: April 2026
			</p>

			<div className="space-y-6">
				<Card>
					<CardHeader>
						<CardTitle className="text-base">
							1. Experimental Nature of This Service
						</CardTitle>
					</CardHeader>
					<CardContent className="space-y-3 text-sm leading-relaxed text-muted-foreground">
						<p>
							JayceTrades is an <strong className="text-foreground">experimental, educational project</strong> that
							generates AI-powered options trade suggestions. The
							platform is provided on an &quot;as-is&quot; basis for
							informational and entertainment purposes only. It is
							not a registered investment advisory service, broker-dealer,
							or financial institution.
						</p>
						<p>
							All trade ideas presented on this platform are
							machine-generated suggestions, not recommendations.
							They are produced by large language models (currently
							GPT-5.4) analyzing publicly available sentiment data
							and market information. These outputs have not been
							reviewed, verified, or endorsed by any licensed
							financial professional.
						</p>
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle className="text-base">
							2. Not Financial Advice
						</CardTitle>
					</CardHeader>
					<CardContent className="space-y-3 text-sm leading-relaxed text-muted-foreground">
						<p>
							<strong className="text-foreground">
								Nothing on JayceTrades constitutes financial,
								investment, tax, or legal advice.
							</strong>{" "}
							The trade suggestions, performance analytics, and
							any other content should not be interpreted as a
							recommendation to buy, sell, or hold any security or
							financial instrument.
						</p>
						<p>
							You should always consult with a qualified, licensed
							financial advisor before making any investment
							decisions. Do not rely on this platform as a
							substitute for professional financial guidance.
						</p>
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle className="text-base">
							3. Significant Risk Disclosure
						</CardTitle>
					</CardHeader>
					<CardContent className="space-y-3 text-sm leading-relaxed text-muted-foreground">
						<p>
							<strong className="text-foreground">
								Options trading involves substantial risk of loss
								and is not suitable for all investors.
							</strong>{" "}
							You can lose your entire investment — and in some
							cases, losses can exceed your initial investment.
							Short-dated options (0-7 DTE), which are the focus
							of this platform, are especially volatile and carry
							elevated risk of total loss.
						</p>
						<p>
							Past performance displayed on this platform — whether
							hypothetical, simulated, or based on actual market
							data — does not guarantee future results. The P&L
							figures shown are estimates based on option mark
							prices at market open and close, and may not reflect
							actual executable prices due to bid-ask spreads,
							liquidity, and market microstructure.
						</p>
						<p>
							By using this platform, you acknowledge that you
							understand these risks and accept full responsibility
							for any trading decisions you make.
						</p>
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle className="text-base">
							4. Hypothetical Performance
						</CardTitle>
					</CardHeader>
					<CardContent className="space-y-3 text-sm leading-relaxed text-muted-foreground">
						<p>
							All performance metrics on JayceTrades are{" "}
							<strong className="text-foreground">hypothetical</strong>.
							They assume that each suggested trade was entered at
							the estimated market open price and exited at the
							closing mark price, with one contract per trade.
							No actual trades are executed by this platform.
						</p>
						<p>
							Hypothetical results have inherent limitations.
							Unlike actual trading, simulated results do not
							account for slippage, commissions, margin
							requirements, the impact of liquidity, or the
							psychological factors of real capital at risk.
						</p>
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle className="text-base">
							5. Data Sources & Accuracy
						</CardTitle>
					</CardHeader>
					<CardContent className="space-y-3 text-sm leading-relaxed text-muted-foreground">
						<p>
							Trade suggestions are generated using data from
							third-party sources including Reddit (r/wallstreetbets,
							r/options), OpenAI&apos;s language models, and the Schwab
							Market Data API. While we strive for accuracy, we
							make no guarantees regarding the completeness,
							reliability, or timeliness of any data presented.
						</p>
						<p>
							Market data, option prices, and stock quotes may be
							delayed or inaccurate. Always verify prices with your
							broker before placing any trades.
						</p>
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle className="text-base">
							6. No Warranty & Limitation of Liability
						</CardTitle>
					</CardHeader>
					<CardContent className="space-y-3 text-sm leading-relaxed text-muted-foreground">
						<p>
							JayceTrades is provided without warranty of any kind,
							express or implied. The creator of this platform shall
							not be held liable for any financial losses, damages,
							or other consequences arising from your use of or
							reliance on the information provided.
						</p>
						<p>
							This platform may experience downtime, data
							inaccuracies, or system errors. Trade suggestions
							may be delayed, missing, or incorrect. Use the
							platform at your own risk.
						</p>
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle className="text-base">
							7. Contact
						</CardTitle>
					</CardHeader>
					<CardContent className="text-sm leading-relaxed text-muted-foreground">
						<p>
							This project is built and maintained by{" "}
							<a
								href="https://jaycebordelon.com"
								target="_blank"
								rel="noopener noreferrer"
								className="font-medium text-primary underline underline-offset-2 hover:text-foreground"
							>
								Jayce Bordelon
							</a>
							. For questions or concerns, reach out via the
							contact information on the personal site.
						</p>
					</CardContent>
				</Card>
			</div>
		</div>
	);
}
