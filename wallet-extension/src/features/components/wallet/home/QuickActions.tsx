import { COLORS, ParagraphSmall } from "@nilfoundation/ui-kit";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import arrowDownIcon from "../../../../../public/icons/arrows/arrow-down.svg";
import arrowUpIcon from "../../../../../public/icons/arrows/arrow-up.svg";
import plusIcon from "../../../../../public/icons/plus.svg";
import { Box, Icon } from "../../shared";

export const QuickActions = () => {
  const { t } = useTranslation("translation");
  const navigate = useNavigate();

  const actions = [
    { icon: plusIcon, label: t("wallet.quickActions.topUp"), link: "/top-up" },
    { icon: arrowUpIcon, label: t("wallet.quickActions.send"), link: "/send" },
    { icon: arrowDownIcon, label: t("wallet.quickActions.receive"), link: "/receive" },
  ];

  return (
    <Box
      $style={{
        display: "flex",
        flexDirection: "row",
        justifyContent: "space-between",
        alignItems: "center",
        width: "100%",
        gap: "8px",
      }}
    >
      {actions.map((action, index) => (
        <Box
          key={`${action.label}-${index}`}
          $align="center"
          $justify="center"
          onClick={() => navigate(action.link)}
          $style={{
            width: "112px",
            height: "78px",
            backgroundColor: COLORS.gray800,
            borderRadius: "8px",
            display: "flex",
            flexDirection: "column",
            ":hover": {
              backgroundColor: COLORS.gray700,
            },
            cursor: "pointer",
          }}
        >
          <Icon
            src={action.icon}
            alt={`${action.label} Icon`}
            size={24}
            iconSize="100%"
            background="transparent"
            pointer={true}
          />
          <ParagraphSmall $style={{ color: COLORS.gray200, marginTop: "8px" }}>
            {action.label}
          </ParagraphSmall>
        </Box>
      ))}
    </Box>
  );
};
