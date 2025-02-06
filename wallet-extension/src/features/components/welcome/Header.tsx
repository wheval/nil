import { COLORS, ParagraphMedium } from "@nilfoundation/ui-kit";
import { useTranslation } from "react-i18next";
import { styled } from "styletron-react";
import extensionIcon from "../../../../public/icons/extension.svg";
import nilIcon from "../../../../public/icons/logo/nil.svg";
import pinIcon from "../../../../public/icons/pin.svg";
import { Box, Icon, Logo } from "../shared";

const PinInstructionBox = styled(Box, {
  marginTop: "12px",
  width: "100%",
  backgroundColor: COLORS.gray900,
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  padding: "12px 16px",
  borderRadius: "8px",
});

export const Header = () => {
  const { t } = useTranslation("translation");
  return (
    <Box
      $align="flex-start"
      $justify="space-between"
      $style={{ width: "100%", padding: "0 48px", flexDirection: "row" }}
    >
      {/* Left Section */}
      <Box $align="flex-start" $justify="flex-start" $style={{ paddingTop: "19px", flex: 1 }}>
        <Logo size={60} />
      </Box>

      {/* Right Section */}
      <Box $align="flex-end" $gap="12px" $style={{ paddingTop: "45px", flex: 3 }}>
        <div>
          <ParagraphMedium>
            {t("welcomePage.header.pinInstruction.mainText")}
            <Icon src={extensionIcon} alt="Extension Icon" size={30} margin={"0px 5px"} />
            {t("welcomePage.header.pinInstruction.walletName")}
          </ParagraphMedium>

          <PinInstructionBox>
            <Box
              $align="center"
              $justify="space-between"
              $style={{
                flexDirection: "row",
              }}
            >
              {/* Left Group */}
              <Box
                $align="center"
                $gap="12px"
                $justify="flex-start"
                $style={{
                  flex: "1 1 auto",
                  flexDirection: "row",
                  overflow: "hidden",
                }}
              >
                <Icon
                  src={nilIcon}
                  alt="Nil Icon"
                  round={false}
                  background={COLORS.black}
                  hoverBackground={COLORS.black}
                />
                <ParagraphMedium style={{ whiteSpace: "nowrap" }}>
                  {t("welcomePage.header.pinInstruction.walletName")}
                </ParagraphMedium>
              </Box>

              <Icon src={pinIcon} alt="Pin Icon" size={30} />
            </Box>
          </PinInstructionBox>
        </div>
      </Box>
    </Box>
  );
};
