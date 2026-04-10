"use client";

import { useState } from "react";
import { TopBar } from "@/components/layout/top-bar";
import { Footer } from "@/components/layout/footer";
import { SubscribeModal } from "@/components/subscribe/subscribe-modal";

export default function AppLayout({ children }: { children: React.ReactNode }) {
	const [modalOpen, setModalOpen] = useState(false);

	return (
		<div className="min-h-dvh">
			<TopBar onSubscribe={() => setModalOpen(true)} />
			{children}
			<Footer />
			<SubscribeModal open={modalOpen} onOpenChange={setModalOpen} />
		</div>
	);
}
