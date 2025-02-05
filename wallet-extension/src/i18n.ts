import i18n from "i18next";
import { initReactI18next } from "react-i18next";

import translationEn from "../public/locales/en/translation.json";
import translationRu from "../public/locales/ru/transalation.json";

const RESOURCES = {
  en: { translation: translationEn },
  ru: { translation: translationRu },
};

export const defaultNS = "translation";

i18n.use(initReactI18next).init({
  resources: RESOURCES,
  defaultNS,
  fallbackLng: "en",
});

export { i18n };
