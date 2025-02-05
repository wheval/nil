import { COLORS, HeadingMedium, ParagraphSmall } from "@nilfoundation/ui-kit";
import { useStore } from "effector-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import ghostIcon from "../../../../../public/icons/ghost.svg";
import { ActivityType } from "../../../../background/storage";
import { $activities } from "../../../store/model/activities.ts";
import { Box, Icon } from "../../shared";

const explorerBaseUrl = import.meta.env.VITE_NIL_EXPLORER || "";

export const ActivityTab = () => {
  const { t } = useTranslation("translation");
  const activities = useStore($activities);

  const activityItems = activities.map((activity) => {
    const isSuccess = activity.success;
    return {
      icon: `/icons/actions/${activity.activityType.toLowerCase()}.svg`,
      hoverIcon: `/icons/actions/${activity.activityType.toLowerCase()}-hover.svg`,
      title: activity.activityType === ActivityType.SEND ? "Sent" : "Topped Up",
      subtitle: isSuccess ? "Confirmed" : "Failed",
      subtitleColor: isSuccess ? COLORS.green200 : COLORS.red300,
      rightText: `${activity.activityType === ActivityType.SEND ? "-" : "+"}${activity.amount} ${activity.token}`,
      rightTextColor:
        activity.activityType === ActivityType.SEND
          ? COLORS.red300
          : isSuccess
            ? COLORS.green200
            : COLORS.red300,
      txHash: activity.txHash,
    };
  });

  if (activityItems.length === 0) {
    return (
      <Box $style={{ textAlign: "center", paddingTop: "40px" }}>
        <Box
          $style={{
            width: "56px",
            height: "56px",
            margin: "0 auto",
            backgroundColor: "transparent",
          }}
        >
          <Icon src={ghostIcon} alt="No Activity" size={56} iconSize="100%" />
        </Box>
        <ParagraphSmall $style={{ marginTop: "16px", color: "inherit" }}>
          {t("wallet.activityTab.noActivityText")}
        </ParagraphSmall>
      </Box>
    );
  }

  return (
    <Box
      $style={{
        paddingTop: "24px",
        display: "flex",
        flexDirection: "column",
        gap: "12px",
        maxHeight: "calc(100vh - 120px)",
        overflowY: "auto",
        "-ms-overflow-style": "none",
        "scrollbar-width": "none",
        height: "285px",
        paddingLeft: "8px",
        paddingRight: "8px",
      }}
    >
      {[...activityItems].reverse().map((activity, index) => {
        const [isHovered, setIsHovered] = useState(false);

        return (
          <Box
            key={`${activity.title}-${index}`}
            $align="center"
            $justify="space-between"
            $style={{
              flexDirection: "row",
              width: "100%",
              padding: "5px",
              cursor: "pointer", // Make the item clickable
              ":hover": {
                backgroundColor: COLORS.gray800,
              },
            }}
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
            onClick={() => {
              const url = `${explorerBaseUrl}tx/${activity.txHash}`;
              window.open(url, "_blank");
            }}
          >
            {/* Left Section: Icon and Title/Subtitle */}
            <Box
              $align="center"
              $gap="8px"
              $style={{
                flexDirection: "row",
              }}
            >
              <Icon
                src={isHovered ? activity.hoverIcon : activity.icon}
                alt={`${activity.title} Icon`}
                size={64}
                iconSize="100%"
                background="transparent"
              />
              <Box $align="flex-start" $style={{ flexDirection: "column" }}>
                <HeadingMedium $style={{ color: COLORS.gray50 }}>{activity.title}</HeadingMedium>
                <ParagraphSmall
                  $style={{
                    color: activity.subtitleColor || COLORS.gray200,
                    whiteSpace: "nowrap",
                  }}
                >
                  {activity.subtitle}
                </ParagraphSmall>
              </Box>
            </Box>

            {/* Right Section: Text */}
            <Box
              $align="center"
              $justify="flex-end"
              $gap="8px"
              $style={{
                flexDirection: "row",
              }}
            >
              <ParagraphSmall
                $style={{
                  color: activity.rightTextColor || COLORS.gray50,
                  textAlign: "right",
                  whiteSpace: "nowrap",
                }}
              >
                {activity.rightText}
              </ParagraphSmall>
            </Box>
          </Box>
        );
      })}
    </Box>
  );
};
