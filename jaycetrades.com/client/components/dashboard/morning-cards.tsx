"use client";

import { ChevronDown } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import {
	Collapsible,
	CollapsibleContent,
	CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Metric } from "@/components/ui/metric";
import {
	calcBreakeven,
	calcMaxGain,
	calcMaxLoss,
	calcMoneyness,
	calcRiskReward,
	sentimentColor,
	sentimentLabel,
} from "@/lib/calculations";
import { fmt, fmtMoney, fmtMoneyInt } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { DashboardTrade, LiveQuotesResponse } from "@/types/trade";

interface MorningCardsProps {
	trades: DashboardTrade[];
	liveQuotes?: LiveQuotesResponse | null;
}

export function MorningCards({ trades, liveQuotes }: MorningCardsProps) {
	return (
		<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{trades.map((dt, index) => (
				<MorningCard
					key={dt.trade.rank}
					dt={dt}
					liveQuotes={liveQuotes}
					index={index}
				/>
			))}
		</div>
	);
}

interface MorningCardProps {
	dt: DashboardTrade;
	liveQuotes?: LiveQuotesResponse | null;
	index: number;
}

function MorningCard({ dt, liveQuotes, index }: MorningCardProps) {
	const [open, setOpen] = useState(false);
	const { trade } = dt;
	const moneyness = calcMoneyness(trade);
	const breakeven = calcBreakeven(trade);
	const maxLoss = calcMaxLoss(trade);
	const maxGain = calcMaxGain(trade);
	const riskReward = calcRiskReward(trade);

	const liveStock = liveQuotes?.quotes?.[trade.symbol];
	const optionKey = Object.keys(liveQuotes?.options ?? {}).find((k) =>
		k.startsWith(trade.symbol),
	);
	const liveOption = optionKey ? liveQuotes?.options?.[optionKey] : null;

	const hasLiveData =
		Boolean(liveQuotes?.connected) && (liveStock || liveOption);

	const liveStockChangeColor = liveStock
		? liveStock.net_change >= 0
			? "text-green"
			: "text-red"
		: "";

	const stockPriceValue = liveStock ? (
		<span
			className={cn(
				"text-sm font-semibold tabular-nums",
				liveStockChangeColor,
			)}
		>
			{fmtMoney(liveStock.last_price)}
			<span className="ml-1 text-xs">
				{liveStock.net_change >= 0 ? "\u2191" : "\u2193"}
			</span>
		</span>
	) : (
		fmtMoney(trade.current_price)
	);

	const riskBadgeVariant: "destructive" | "outline" | "secondary" =
		trade.risk_level === "HIGH"
			? "destructive"
			: trade.risk_level === "MEDIUM"
				? "outline"
				: "secondary";

	return (
		<Card
			className="group animate-in fade-in slide-in-from-bottom-1 duration-300 transition-all hover:-translate-y-0.5 hover:shadow-md"
			style={{ animationDelay: `${index * 60}ms` }}
		>
			<CardContent className="space-y-4 p-5">
				{/* Tier 1: header */}
				<div className="flex flex-wrap items-center gap-1.5">
					<Badge variant="secondary">#{trade.rank}</Badge>
					<span className="text-xl font-bold tracking-tight">
						${trade.symbol}
					</span>
					<Badge
						variant="outline"
						className={cn(
							trade.contract_type === "CALL"
								? "border-green-border text-green"
								: "border-red-border text-red",
						)}
					>
						{trade.contract_type}
					</Badge>
					<Badge variant={moneyness.variant}>{moneyness.label}</Badge>
					<Badge variant={riskBadgeVariant}>{trade.risk_level}</Badge>
				</div>

				{/* Tier 1: premium */}
				<div>
					<div className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
						Est. Premium
					</div>
					<div className="mt-1 text-[32px] font-semibold tabular-nums leading-none">
						{fmtMoney(trade.estimated_price)}
					</div>
					<div className="mt-1 text-xs text-muted-foreground">
						{fmtMoneyInt(trade.estimated_price * 100)} per contract
					</div>
				</div>

				{/* Tier 2: key metrics */}
				<div className="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
					<Metric label="Strike" value={fmtMoney(trade.strike_price)} />
					<Metric
						label="Expiration"
						value={`${trade.expiration} (${trade.dte}d)`}
					/>
					<Metric label="Stock Price" value={stockPriceValue} />
					<Metric label="Target" value={fmtMoney(trade.target_price)} />
				</div>

				{/* Tier 2: catalyst */}
				{trade.catalyst && (
					<div className="rounded-md bg-amber-bg px-3 py-2 text-sm">
						<span className="font-semibold text-amber">Catalyst:</span>{" "}
						{trade.catalyst}
					</div>
				)}

				{/* Tier 3: thesis preview */}
				{trade.thesis && (
					<p className="line-clamp-2 text-sm leading-relaxed text-muted-foreground">
						{trade.thesis}
					</p>
				)}

				{/* Tier 4: Collapsible details */}
				<Collapsible open={open} onOpenChange={setOpen}>
					<CollapsibleTrigger asChild>
						<button
							type="button"
							className="group/details flex w-full items-center justify-between rounded-md border bg-muted/40 px-3 py-2 text-xs font-medium text-muted-foreground transition-colors hover:bg-muted"
						>
							<span>View full details</span>
							<ChevronDown className="h-3.5 w-3.5 transition-transform group-data-[state=open]/details:rotate-180" />
						</button>
					</CollapsibleTrigger>
					<CollapsibleContent className="overflow-hidden data-[state=closed]:animate-collapsible-up data-[state=open]:animate-collapsible-down">
						<div className="space-y-4 pt-4">
							<div className="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
								<Metric
									label="Breakeven"
									value={fmtMoney(breakeven)}
								/>
								<Metric
									label="Max Loss"
									value={
										<span className="text-sm font-semibold tabular-nums text-red">
											{fmtMoneyInt(maxLoss)}
										</span>
									}
								/>
								{maxGain !== null && (
									<Metric
										label="Target Gain"
										value={
											<span className="text-sm font-semibold tabular-nums text-green">
												{fmtMoneyInt(maxGain)}
											</span>
										}
									/>
								)}
								<Metric
									label="Risk / Reward"
									value={
										riskReward
											? `1:${riskReward.toFixed(1)}`
											: "N/A"
									}
								/>
								<Metric
									label="Sentiment"
									value={
										<span
											className={cn(
												"text-sm font-semibold tabular-nums",
												sentimentColor(trade.sentiment_score),
											)}
										>
											{sentimentLabel(trade.sentiment_score)} (
											{fmt(trade.sentiment_score, 2)})
										</span>
									}
								/>
								<Metric
									label="Stop Loss"
									value={fmtMoney(trade.stop_loss)}
								/>
							</div>

							{trade.thesis && (
								<p className="text-sm leading-relaxed text-muted-foreground">
									{trade.thesis}
								</p>
							)}

							{hasLiveData && (
								<div className="rounded-md border border-green-border bg-green-bg p-3">
									<div className="text-[11px] font-semibold uppercase tracking-wider text-green">
										Live Data
									</div>
									<div className="mt-2 grid grid-cols-2 gap-2 text-sm">
										{liveStock && (
											<Metric
												label="Stock"
												value={fmtMoney(liveStock.last_price)}
											/>
										)}
										{liveStock && (
											<Metric
												label="Change"
												value={
													<span
														className={cn(
															"text-sm font-semibold tabular-nums",
															liveStockChangeColor,
														)}
													>
														{fmt(liveStock.net_change_pct, 2)}%
													</span>
												}
											/>
										)}
										{liveOption && (
											<Metric
												label="Option Mark"
												value={fmtMoney(liveOption.mark)}
											/>
										)}
										{liveOption && (
											<Metric
												label="Delta"
												value={fmt(liveOption.delta, 2)}
											/>
										)}
									</div>
								</div>
							)}
						</div>
					</CollapsibleContent>
				</Collapsible>
			</CardContent>
		</Card>
	);
}
