import { ParagraphSmall } from "@nilfoundation/ui-kit";
import { useTranslation } from "react-i18next";
import nftIcon from "../../../../../public/icons/in-progress.svg";
import { Box, Icon } from "../../shared";

export const NFTTab = () => {
  const { t } = useTranslation("translation");

  return (
    <Box $style={{ textAlign: "center", paddingTop: "40px" }}>
      {/* Icon */}
      <Box
        $style={{
          width: "56px",
          height: "56px",
          margin: "0 auto",
          backgroundColor: "transparent",
        }}
      >
        <Icon src={nftIcon} alt="NFT Placeholder" size={56} iconSize="100%" />
      </Box>

      {/* Text */}
      <ParagraphSmall $style={{ marginTop: "16px", color: "inherit" }}>
        {t("wallet.nftTab.placeholder.text")}{" "}
        <a
          href="https://t.me/NilDevnetTokenBot"
          target="_blank"
          rel="noopener noreferrer"
          style={{ textDecoration: "underline" }}
        >
          {t("wallet.nftTab.placeholder.linkText")}
        </a>
      </ParagraphSmall>
    </Box>
  );
};
