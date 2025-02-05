import {
  Card as BaseCard,
  Button,
  COLORS,
  HeadingLarge,
  ParagraphMedium,
} from "@nilfoundation/ui-kit";
import { useTranslation } from "react-i18next";
import openPopupCommandMac from "../../../../public/img/open-popup-mac.png";
import openPopupCommandWin from "../../../../public/img/open-popup-win.png";
import welcomeCardHeader from "../../../../public/img/welcome-card-header.png";
import { getOperatingSystem } from "../../utils";
import { Box, Image } from "../shared";

export const Card = () => {
  const { t } = useTranslation("translation");
  const handleOpenPopup = async () => {
    if (chrome?.action?.openPopup) {
      await chrome.action.openPopup();
    } else {
      console.error("Popup action is not available");
    }
  };

  const operatingSystem = getOperatingSystem();
  const imageSrc = operatingSystem === "mac" ? openPopupCommandMac : openPopupCommandWin;

  return (
    <Box $position="absolute" $top="50%" $left="50%" $transform="translate(-50%, -50%)">
      <BaseCard
        overrides={{
          Root: {
            style: ({ $theme }) => ({
              paddingLeft: "24px",
              paddingRight: "24px",
              paddingTop: "0px",
              paddingBottom: "8px",
              backgroundColor: COLORS.gray900,
              textAlign: "center",
              width: "90%",
              maxWidth: "500px",
              height: "auto",
              display: "flex",
              flexDirection: "column",
              justifyContent: "center",
              alignItems: "center",
              borderRadius: "8px",
              boxShadow: $theme.lighting.shadow600,
              "@media (max-width: 768px)": {
                width: "80%",
              },
              "@media (max-width: 550px)": {
                width: "95%",
              },
            }),
          },
        }}
      >
        {/* Card Header Image */}
        <Image src={welcomeCardHeader} alt="Card Header" draggable="false" />

        {/* Text Section */}
        <Box $padding="0 24px" $gap="24px" $align="center" $style={{ marginBottom: "24px" }}>
          <HeadingLarge>{t("welcomePage.card.heading")}</HeadingLarge>
          <ParagraphMedium $style={{ color: COLORS.gray300 }}>
            {t("welcomePage.card.description")}
          </ParagraphMedium>
        </Box>

        {/* Card Command Image */}
        <Image src={imageSrc} alt="Card Command" draggable="false" />

        {/* Button - Only show if openPopup is available */}
        {chrome?.action?.openPopup && (
          <Button
            overrides={{
              Root: {
                style: { width: "100%", height: "48px" },
              },
            }}
            onClick={handleOpenPopup}
          >
            {t("welcomePage.card.buttonLabel")}
          </Button>
        )}
      </BaseCard>
    </Box>
  );
};
