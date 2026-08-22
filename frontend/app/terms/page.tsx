import { getTranslations } from "next-intl/server";

import { LegalPage, type LegalSection } from "@/components/legal-page";

const sectionKeys = ["s1", "s2", "s3", "s4", "s5", "s6", "s7"] as const;

export default async function TermsPage() {
  const t = await getTranslations("legal.terms");
  const common = await getTranslations("legal");

  const sections: LegalSection[] = sectionKeys.map((key) => ({
    title: t(`${key}.title`),
    body: t(`${key}.body`),
  }));

  return (
    <LegalPage
      title={t("title")}
      updated={common("updated")}
      sections={sections}
      backLabel={common("back")}
    />
  );
}
