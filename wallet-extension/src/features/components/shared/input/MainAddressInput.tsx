import { COLORS, Input, ParagraphMedium } from "@nilfoundation/ui-kit";
import { useStore } from "effector-react";
import { useTranslation } from "react-i18next";
import walletIcon from "../../../../../public/icons/wallet.svg";
import { $smartAccount } from "../../../store/model/smartAccount.ts";
import { formatAddress } from "../../../utils";
import { Box, Icon } from "../index.ts";

export const MainAddressInput = () => {
  const { t } = useTranslation("translation");

  // Get the smartAccount from the store
  const smartAccount = useStore($smartAccount);

  return (
    <Input
      overrides={{
        Root: {
          style: {
            width: "100%",
          },
        },
        InputContainer: {
          style: {
            color: COLORS.gray50,
            paddingLeft: "12px",
          },
        },
      }}
      startEnhancer={() => (
        <Box
          $align="center"
          $gap="8px"
          $style={{
            flexDirection: "row",
          }}
        >
          {/* Icon */}
          <Icon
            src={walletIcon}
            alt="Wallet Icon"
            size={24}
            iconSize="100%"
            background="transparent"
          />

          {/* Label */}
          <ParagraphMedium
            style={{
              color: COLORS.gray50,
              whiteSpace: "nowrap",
            }}
          >
            {t("wallet.header.mainWallet")}
          </ParagraphMedium>

          {/* Address */}
          <ParagraphMedium style={{ color: COLORS.gray200 }}>
            {smartAccount ? formatAddress(smartAccount.address) : ""}
          </ParagraphMedium>
        </Box>
      )}
    />
  );
};
