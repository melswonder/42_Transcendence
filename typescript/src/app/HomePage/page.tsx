'use client';

import React from "react";
import API from "../api/page";

export default function HomePage() {
  return (
    <div className="min-h-screen p-8">
      <h1 className="text-3xl font-bold mb-6">ホームページ</h1>
      <API />
    </div>
  );
}